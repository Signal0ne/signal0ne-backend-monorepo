"""A module to perform Google Custom Search API queries
and fetch summaries of the search results."""
import os
import json
from newspaper.configuration import Configuration
from transformers import pipeline, AutoTokenizer, TFAutoModelForSeq2SeqLM
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
        self.num_results = 3
        self.tokenizer = AutoTokenizer.from_pretrained('VidhuMathur/bart-log-summarization')
        self.model = TFAutoModelForSeq2SeqLM.from_pretrained('VidhuMathur/bart-log-summarization')
        self.model_pipeline = pipeline("summarization", model=self.model, tokenizer=self.tokenizer)
    

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
            config = Configuration()
            config.browser_user_agent = ('Mozilla/5.0 (Macintosh; Intel Mac OS X 11_1_0) '
                                         'AppleWebKit/537.36 (KHTML, like Gecko) '
                                         'Chrome/97.0.4692.99 Safari/537.36')
            article = Article(url, config=config)
            article.download()
            article.parse()
            article.nlp()
            
            summary = self.model_pipeline(article.text, max_length=500, min_length=200, do_sample=False)
            return summary[0]['summary_text']
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
    
if __name__ == "__main__":
        search = GoogleCustomSearch()
        queries = {
            "queries": [
                {"question": "MySQL Server Failure: Corrupted InnoDB Plugin and Datafile Issues"},
            ]
        }
        results = search.run_search(queries)
        print(results)
