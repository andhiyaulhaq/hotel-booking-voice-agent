import os
from pathlib import Path
from langchain_community.document_loaders import TextLoader
from langchain_text_splitters import RecursiveCharacterTextSplitter
from langchain_openai import OpenAIEmbeddings
from langchain_community.vectorstores import FAISS

# We'll use a local directory to store/load the FAISS index
INDEX_DIR = Path(__file__).parent.parent / "faiss_index"
DATA_FILE = Path(__file__).parent.parent / "data" / "hotel_policies.txt"

def get_retriever():
    """
    Initializes and returns a FAISS retriever for the hotel policies.
    If the index doesn't exist on disk, it reads the text file, embeds it, and saves the index.
    """
    embeddings = OpenAIEmbeddings()
    
    if INDEX_DIR.exists():
        # Load existing index
        print("Loading existing FAISS index from disk...")
        vectorstore = FAISS.load_local(str(INDEX_DIR), embeddings, allow_dangerous_deserialization=True)
    else:
        # Create new index
        print("Creating new FAISS index from text file...")
        loader = TextLoader(str(DATA_FILE))
        docs = loader.load()
        
        text_splitter = RecursiveCharacterTextSplitter(
            chunk_size=500,
            chunk_overlap=50,
            separators=["\n\n", "\n", ".", " ", ""]
        )
        splits = text_splitter.split_documents(docs)
        
        vectorstore = FAISS.from_documents(splits, embeddings)
        
        # Ensure directory exists and save
        INDEX_DIR.mkdir(parents=True, exist_ok=True)
        vectorstore.save_local(str(INDEX_DIR))
        print("Saved FAISS index to disk.")
        
    return vectorstore.as_retriever(search_kwargs={"k": 3})

# Expose a simple function for the agent tool to use
retriever = get_retriever()

def search_hotel_policies(query: str) -> str:
    """Search the hotel policies for relevant information to answer the guest's question."""
    docs = retriever.invoke(query)
    if not docs:
        return "I'm sorry, I don't have information on that in my knowledge base."
    
    # Combine the retrieved contexts
    context = "\n\n".join([doc.page_content for doc in docs])
    return context
