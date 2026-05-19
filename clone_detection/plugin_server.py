#!/usr/bin/env python3
"""
HTTP Plugin Server for Semantic Clone Detection

This server provides an HTTP API for the semantic clone detection functionality,
allowing it to be used as an optional plugin for structurelint.

The plugin architecture keeps the core binary small while providing advanced
ML-based clone detection as an opt-in feature.
"""

import logging
import time
from pathlib import Path
from threading import Lock
from typing import List, Optional

from fastapi import FastAPI, HTTPException
from fastapi.responses import JSONResponse
from pydantic import BaseModel, Field
import uvicorn

# Import clone detection modules (these require heavy dependencies)
try:
    from clone_detection.embeddings.graphcodebert import GraphCodeBERTEmbedder
    from clone_detection.indexing.faiss_index import FAISSIndexBuilder, IndexType
    from clone_detection.parsers.tree_sitter_parser import TreeSitterParser
    from clone_detection.query.metadata import MetadataStore
    from clone_detection.query.search import CloneSearcher
    DEPENDENCIES_AVAILABLE = True
except ImportError as e:
    DEPENDENCIES_AVAILABLE = False
    IMPORT_ERROR = str(e)

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s"
)
logger = logging.getLogger(__name__)

# Create FastAPI app
app = FastAPI(
    title="Structurelint Semantic Clone Detection Plugin",
    version="0.1.0",
    description="HTTP API for semantic code clone detection using GraphCodeBERT"
)

# Global state with thread-safe initialization
embedder: Optional[GraphCodeBERTEmbedder] = None
embedder_lock = Lock()  # Thread-safe initialization


# --- API Models ---

class SemanticCloneRequest(BaseModel):
    """Request model for semantic clone detection"""
    source_dir: str = Field(..., description="Root directory to analyze")
    languages: Optional[List[str]] = Field(
        default=["python", "go", "javascript"],
        description="Languages to analyze"
    )
    exclude_patterns: Optional[List[str]] = Field(
        default=["**/*_test.*", "**/node_modules/**", "**/vendor/**"],
        description="Glob patterns to exclude"
    )
    similarity_threshold: float = Field(
        default=0.85,
        ge=0.0,
        le=1.0,
        description="Similarity threshold (0.0-1.0)"
    )
    max_results: int = Field(
        default=100,
        ge=1,
        description="Maximum number of clone pairs to return"
    )


class SemanticClone(BaseModel):
    """Model for a detected semantic clone pair"""
    source_file: str
    source_start_line: int
    source_end_line: int
    target_file: str
    target_start_line: int
    target_end_line: int
    similarity: float
    explanation: Optional[str] = None


class SemanticCloneStats(BaseModel):
    """Statistics about the clone detection analysis"""
    files_analyzed: int
    functions_analyzed: int
    duration_ms: int
    model_used: str = "microsoft/graphcodebert-base"


class SemanticCloneResponse(BaseModel):
    """Response model for semantic clone detection"""
    clones: List[SemanticClone]
    stats: SemanticCloneStats
    error: Optional[str] = None


class HealthResponse(BaseModel):
    """Health check response"""
    status: str  # "healthy", "degraded", "unhealthy"
    version: str = "0.1.0"
    capabilities: List[str] = [
        "semantic-clone-detection",
        "graphcodebert-embeddings",
        "faiss-indexing"
    ]
    message: Optional[str] = None


# --- API Endpoints ---

@app.get("/health")
async def health_check() -> HealthResponse:
    """
    Health check endpoint.

    Returns the plugin's health status and capabilities.
    """
    if not DEPENDENCIES_AVAILABLE:
        return HealthResponse(
            status="unhealthy",
            capabilities=[],
            message=f"Dependencies not available: {IMPORT_ERROR}"
        )

    # Check if model can be loaded (thread-safe)
    global embedder
    if embedder is None:
        with embedder_lock:
            # Double-check after acquiring lock
            if embedder is None:
                try:
                    # Try to initialize embedder (this can fail if model not downloaded)
                    embedder = GraphCodeBERTEmbedder(
                        model_name="microsoft/graphcodebert-base",
                        device="cpu"  # Use CPU for health check
                    )
                    return HealthResponse(
                        status="healthy",
                        message="Semantic clone detection ready"
                    )
                except Exception as e:
                    return HealthResponse(
                        status="degraded",
                        message=f"Model loading failed: {str(e)}",
                        capabilities=["limited"]
                    )

    return HealthResponse(
        status="healthy",
        message="Semantic clone detection ready"
    )


@app.post("/api/v1/detect")
async def detect_clones(request: SemanticCloneRequest) -> SemanticCloneResponse:
    """
    Detect semantic clones in the specified source directory.

    This endpoint analyzes code using GraphCodeBERT embeddings and FAISS
    similarity search to find semantically similar code fragments.
    """
    if not DEPENDENCIES_AVAILABLE:
        raise HTTPException(
            status_code=503,
            detail=f"Dependencies not available: {IMPORT_ERROR}"
        )

    start_time = time.time()

    try:
        # Initialize components (thread-safe)
        global embedder
        if embedder is None:
            with embedder_lock:
                # Double-check after acquiring lock
                if embedder is None:
                    logger.info("Initializing GraphCodeBERT embedder...")
                    embedder = GraphCodeBERTEmbedder(
                        model_name="microsoft/graphcodebert-base",
                        device="cpu"  # TODO: Support GPU via config
                    )

        # Validate and sanitize source directory to prevent path traversal
        source_path = Path(request.source_dir).resolve()
        if not source_path.exists():
            raise HTTPException(
                status_code=400,
                detail=f"Source directory does not exist: {request.source_dir}"
            )
        if not source_path.is_dir():
            raise HTTPException(
                status_code=400,
                detail=f"Path is not a directory: {request.source_dir}"
            )

        # Parse source files
        logger.info(f"Parsing source directory: {source_path}")
        parser = TreeSitterParser(languages=request.languages)

        functions = parser.parse_directory(
            str(source_path),
            exclude_patterns=request.exclude_patterns
        )
        logger.info(f"Found {len(functions)} functions across {len(set(f.file_path for f in functions))} files")

        if len(functions) == 0:
            return SemanticCloneResponse(
                clones=[],
                stats=SemanticCloneStats(
                    files_analyzed=0,
                    functions_analyzed=0,
                    duration_ms=int((time.time() - start_time) * 1000)
                ),
                error="No functions found in source directory"
            )

        # Generate embeddings
        logger.info("Generating embeddings...")
        embeddings = []
        for func in functions:
            embedding = embedder.embed_single(func.code)
            embeddings.append(embedding)

        # Build FAISS index
        logger.info("Building FAISS index...")
        index_builder = FAISSIndexBuilder(
            dimension=embeddings[0].shape[0],
            index_type=IndexType.FLAT  # Use flat index for accuracy
        )

        import numpy as np
        embeddings_array = np.vstack(embeddings)
        index = index_builder.build(embeddings_array)

        # Search for similar code
        logger.info("Searching for semantic clones...")
        clones = []
        seen_pairs = set()

        for i, query_embedding in enumerate(embeddings):
            # Search for top-k similar functions
            distances, indices = index.search(
                query_embedding.reshape(1, -1),
                k=min(10, len(functions))
            )

            for distance, idx in zip(distances[0], indices[0]):
                # Convert distance to similarity (cosine similarity)
                similarity = 1.0 - (distance / 2.0)  # Normalize to [0, 1]

                # Skip self-matches and below threshold
                if idx == i or similarity < request.similarity_threshold:
                    continue

                # Skip if we've already seen this pair
                pair_key = tuple(sorted([i, idx]))
                if pair_key in seen_pairs:
                    continue
                seen_pairs.add(pair_key)

                # Create clone pair
                source_func = functions[i]
                target_func = functions[idx]

                clone = SemanticClone(
                    source_file=str(source_func.file_path),
                    source_start_line=source_func.start_line,
                    source_end_line=source_func.end_line,
                    target_file=str(target_func.file_path),
                    target_start_line=target_func.start_line,
                    target_end_line=target_func.end_line,
                    similarity=float(similarity),
                    explanation=f"Semantic similarity: {similarity:.3f}"
                )
                clones.append(clone)

                # Stop if we've hit max results
                if len(clones) >= request.max_results:
                    break

            if len(clones) >= request.max_results:
                break

        duration_ms = int((time.time() - start_time) * 1000)

        logger.info(f"Found {len(clones)} semantic clones in {duration_ms}ms")

        return SemanticCloneResponse(
            clones=clones,
            stats=SemanticCloneStats(
                files_analyzed=len(set(f.file_path for f in functions)),
                functions_analyzed=len(functions),
                duration_ms=duration_ms
            )
        )

    except Exception as e:
        logger.error(f"Clone detection failed: {e}", exc_info=True)
        duration_ms = int((time.time() - start_time) * 1000)

        return SemanticCloneResponse(
            clones=[],
            stats=SemanticCloneStats(
                files_analyzed=0,
                functions_analyzed=0,
                duration_ms=duration_ms
            ),
            error=str(e)
        )


@app.get("/")
async def root():
    """Root endpoint with plugin information"""
    return {
        "name": "Structurelint Semantic Clone Detection Plugin",
        "version": "0.1.0",
        "status": "running",
        "endpoints": {
            "health": "/health",
            "detect": "/api/v1/detect"
        }
    }


def main():
    """Run the plugin server"""
    import argparse

    parser = argparse.ArgumentParser(
        description="Structurelint Semantic Clone Detection Plugin Server"
    )
    parser.add_argument(
        "--host",
        default="127.0.0.1",
        help="Host to bind to (default: 127.0.0.1)"
    )
    parser.add_argument(
        "--port",
        type=int,
        default=8765,
        help="Port to bind to (default: 8765)"
    )
    parser.add_argument(
        "--reload",
        action="store_true",
        help="Enable auto-reload for development"
    )

    args = parser.parse_args()

    logger.info(f"Starting plugin server on {args.host}:{args.port}")
    logger.info("Dependencies available: %s", DEPENDENCIES_AVAILABLE)

    if not DEPENDENCIES_AVAILABLE:
        logger.warning("ML dependencies not available. Install with: pip install -e .")

    uvicorn.run(
        "clone_detection.plugin_server:app",
        host=args.host,
        port=args.port,
        reload=args.reload,
        log_level="info"
    )


if __name__ == "__main__":
    main()
