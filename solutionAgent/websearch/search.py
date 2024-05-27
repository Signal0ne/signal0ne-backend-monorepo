"""A module to perform Google Custom Search API queries
and fetch summaries of the search results."""
import os
import json
import requests
import dotenv
from websearch.scrape import WebScraper
import nltk
import torch

class GoogleCustomSearch:
    """A class to perform Google Custom Search API queries and
    fetch summaries of the search results."""
    def __init__(self, model,tokenizer):
        nltk.download('punkt')
        dotenv.load_dotenv()
        api_key = os.getenv('GOOGLE_API_KEY')
        cse_id = os.getenv('GOOGLE_CSE_ID')
        self.api_key = api_key
        self.cse_id = cse_id
        self.base_url = "https://www.googleapis.com/customsearch/v1"
        self.num_results = 3
        self.model = model
        self.tokenizer = tokenizer

    def generate_summary(self, text, max_input_length=1024, max_output_length=512):
        """Generate a summary of the given text using the model."""
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
            outputs =self.model.generate(
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
        response = requests.get(self.base_url, params=payload, timeout=10)
        return response.json()

    def fetch_summary(self, url):
        """Fetch the summary of an article from the given URL."""
        try:
            webScrape = WebScraper(url)
            text = webScrape.get_text()
            summary = self.generate_summary(text)
            return summary
        except Exception as e:
            print(f"Error fetching summary: {e}")
            return ""
            
    def run_search(self, queries):
        """Run a search query and return the search results with summaries."""
        results = []
        global_index = 1
        for query in queries['queries']:
            query = query['question']
            data = self.search(query)
            try:
                if 'items' not in data:
                    query = data['spelling']['correctedQuery']
                    data = self.search(query)
                for item in data['items']:
                    url = item['link']
                    snippet = item['snippet']
                    summary = self.fetch_summary(url)
                    results.append({'index': global_index, 'url': url, 'snippet': snippet, 'summary': summary})
                    global_index += 1
            except Exception as e:
                print(f"Error processing: {data}")
                print(f"Error processing search results: {e}")
                results.append({'index': global_index, 'url': "", 'snippet': "", 'summary': ""})
                global_index += 1
        return json.dumps(results, indent=4)