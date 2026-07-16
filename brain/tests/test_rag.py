import sys
from pathlib import Path
import pytest

# Add src to path
sys.path.append(str(Path(__file__).parent.parent / "src"))
import rag

def test_search_hotel_policies(mocker):
    # Mock the retriever.invoke method
    mock_doc = mocker.Mock()
    mock_doc.page_content = "Pets are allowed with an extra fee."
    
    mock_retriever = mocker.Mock()
    mock_retriever.invoke.return_value = [mock_doc]
    
    # Patch the retriever in rag module
    mocker.patch('rag.retriever', mock_retriever)
    
    # Run function
    result = rag.search_hotel_policies("pets allowed?")
    
    assert "Pets are allowed" in result
    mock_retriever.invoke.assert_called_once_with("pets allowed?")

def test_search_hotel_policies_no_results(mocker):
    mock_retriever = mocker.Mock()
    mock_retriever.invoke.return_value = []
    
    mocker.patch('rag.retriever', mock_retriever)
    
    result = rag.search_hotel_policies("aliens allowed?")
    
    assert "I'm sorry, I don't have information" in result
