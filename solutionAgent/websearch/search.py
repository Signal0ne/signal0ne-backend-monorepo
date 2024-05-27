from multiprocessing import Manager, Value, Lock
import os
import json
import requests
import dotenv
from websearch.scrape import WebScraper
import nltk
import torch
from concurrent.futures import ThreadPoolExecutor, as_completed

class GoogleCustomSearch:
    """A class to perform Google Custom Search API queries and
    fetch summaries of the search results."""
    def __init__(self, model, tokenizer):
        nltk.download('punkt')
        dotenv.load_dotenv()
        self.api_key = os.getenv('GOOGLE_API_KEY')
        self.cse_id = os.getenv('GOOGLE_CSE_ID')
        self.base_url = "https://www.googleapis.com/customsearch/v1"
        self.num_results = 3
        self.model = model
        self.tokenizer = tokenizer
        self.session = requests.Session()  # Create a session

    def __del__(self):
        self.session.close()  # Ensure the session is closed when the object is destroyed

    def generate_summary(self, text, max_input_length=1024, max_output_length=512):
        """Generate a summary of the given text using the model."""
        print("Generating summary...")
        device = "cuda" if torch.cuda.is_available() else "cpu"
        inputs = self.tokenizer(
            text,
            max_length=max_input_length,
            truncation=True,
            padding="max_length",
            return_tensors="pt"
        )
        input_ids = inputs.input_ids.to(device)
        attention_mask = inputs.attention_mask.to(device)

        with torch.no_grad():
            outputs = self.model.generate(
                input_ids=input_ids,
                attention_mask=attention_mask,
                max_length=max_output_length,
                num_beams=4,
                early_stopping=True
            )
        summary = self.tokenizer.decode(outputs[0], skip_special_tokens=True)
        return summary   

    def build_payload(self, query, **kwargs):
        """Build the payload for the Google Custom Search API query."""
        payload = {
            'q': query,
            'key': self.api_key,
            'cx': self.cse_id,
            'num': self.num_results,
        }
        payload.update(kwargs)
        return payload
    
    def search(self, query, **kwargs):
        """Perform a Google Custom Search API query and return the results."""
        payload = self.build_payload(query, **kwargs)
        response = self.session.get(self.base_url, params=payload, timeout=10)
        return response.json()

    def fetch_summary(self, url):
        """Fetch the summary of an article from the given URL."""
        try:
            webScrape = WebScraper(url,self.session)
            text = webScrape.get_text()
            summary = self.generate_summary(text)
            return summary
        except Exception as e:
            print(f"Error fetching summary: {e}")
            return ""
            
    def run_search(self, queries):
        """Run a search query and return the search results with summaries."""
        results = []
        global_index = Value('i', 1)  # Use Value for shared index
        lock = Lock()  # Use Lock to synchronize access to the shared index

        def process_query(query):
            """Helper function to process a single query."""
            local_results = []
            try:
                data = self.search(query)
                if 'items' not in data:
                    query = data.get('spelling', {}).get('correctedQuery', query)
                    data = self.search(query)
                for item in data.get('items', []):
                    url = item['link']
                    snippet = item['snippet']
                    summary = self.fetch_summary(url)
                    with lock:
                        index = global_index.value
                        global_index.value += 1
                    local_results.append({'index': index, 'url': url, 'snippet': snippet, 'summary': summary})
            except Exception as e:
                print(f"Error processing search results: {e}")
                with lock:
                    index = global_index.value
                    global_index.value += 1
                local_results.append({'index': index, 'url': "", 'snippet': "", 'summary': ""})
            return local_results

        with ThreadPoolExecutor(max_workers=10) as executor:
            futures = [executor.submit(process_query, query['question']) for query in queries['queries']]
            for future in as_completed(futures):
                results.extend(future.result())

        # Sort results by their index to maintain the original order
        results.sort(key=lambda x: x['index'])
        return json.dumps(results, indent=4)
