import os
import json
import requests
import dotenv
from newspaper import Article
import nltk

class GoogleCustomSearch:
    """A class to perform Google Custom Search API queries and
    fetch summaries of the search results."""
    def __init__(self):
        nltk.download('punkt')
        dotenv.load_dotenv()
        api_key = os.getenv('GOOGLE_API_KEY')
        cse_id = os.getenv('GOOGLE_CSE_ID')
        self.api_key = api_key
        self.cse_id = cse_id
        self.base_url = "https://www.googleapis.com/customsearch/v1"

    def build_payload(self, query, num_results=3, **kwargs):
        """Build the payload for the Google Custom Search API query."""
        payload = {
            'q': query,
            'key': self.api_key,
            'cx': self.cse_id,
            'num': num_results
        }
        payload.update(kwargs)
        return payload

    def search(self, query, **kwargs):
        """Perform a Google Custom Search API query and return the results."""
        payload = self.build_payload(query, **kwargs)
        response = requests.get(self.base_url, params=payload)
        return response.json()

    def fetch_summary(self, url):
        """Fetch the summary of an article from the given URL."""
        article = Article(url)
        article.download()
        article.parse()
        article.nlp()
        return article.summary

    def run_search(self, query, num_results=3):
        """Run a search query and return the search results with summaries."""
        data = self.search(query, num_results=num_results)
        results = []
        for item in data['items']:
            url = item['link']
            snippet = item['snippet']
            summary = self.fetch_summary(url)
            results.append({'url': url, 'snippet': snippet, 'summary': summary})
        return json.dumps(results, indent=4)
