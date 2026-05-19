"""
Basic tests for the clone detection system.

These tests verify that the core components can be imported and instantiated.
"""

import pytest
import numpy as np


def test_imports():
    """Test that all main components can be imported."""
    from clone_detection.parsers import TreeSitterParser
    from clone_detection.embeddings import GraphCodeBERTEmbedder
    from clone_detection.indexing import FAISSIndexBuilder
    from clone_detection.query import CloneSearcher, MetadataStore

    assert TreeSitterParser is not None
    assert GraphCodeBERTEmbedder is not None
    assert FAISSIndexBuilder is not None
    assert CloneSearcher is not None
    assert MetadataStore is not None


def test_tree_sitter_parser_init():
    """Test TreeSitterParser initialization."""
    from clone_detection.parsers import TreeSitterParser

    parser = TreeSitterParser(languages=["python"])
    assert "python" in parser.get_supported_languages()


def test_cosine_to_l2_conversion():
    """Test the critical cosine-to-L2 threshold conversion."""
    from clone_detection.indexing.faiss_index import (
        cosine_to_l2_threshold,
        l2_to_cosine_similarity,
    )

    # Test the conversion formula from Table 4.1
    test_cases = [
        (0.995, 0.100),
        (0.980, 0.200),
        (0.950, 0.316),
        (0.900, 0.447),
    ]

    for cosine_sim, expected_l2 in test_cases:
        l2_dist = cosine_to_l2_threshold(cosine_sim)
        assert abs(l2_dist - expected_l2) < 0.001, f"Failed for cosine={cosine_sim}"

        # Test round-trip conversion
        recovered_cosine = l2_to_cosine_similarity(l2_dist)
        assert abs(recovered_cosine - cosine_sim) < 0.001


def test_l2_normalization():
    """Test that L2 normalization works correctly."""
    import faiss

    # Create random vectors
    vectors = np.random.randn(10, 768).astype(np.float32)

    # Normalize
    faiss.normalize_L2(vectors)

    # Check that all vectors have unit length
    norms = np.linalg.norm(vectors, axis=1)
    assert np.allclose(norms, 1.0), "Vectors should have unit length after normalization"


def test_code_snippet_dataclass():
    """Test CodeSnippet data structure."""
    from clone_detection.parsers import CodeSnippet

    snippet = CodeSnippet(
        code="def foo(): pass",
        file_path="/test/file.py",
        start_line=1,
        end_line=1,
        language="python",
        function_name="foo",
    )

    assert snippet.code == "def foo(): pass"
    assert snippet.function_name == "foo"
    assert snippet.language == "python"

    # Test to_dict
    data = snippet.to_dict()
    assert data["code"] == "def foo(): pass"
    assert data["function_name"] == "foo"


def test_config_loading():
    """Test configuration loading."""
    from clone_detection.config import CloneDetectionConfig

    # Test default config
    config = CloneDetectionConfig()
    assert config.model.name == "microsoft/graphcodebert-base"
    assert config.index.dimension == 768
    assert config.query.default_similarity == pytest.approx(0.95)


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
