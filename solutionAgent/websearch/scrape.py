"""Module for scraping the text content of a web page."""
import os
import requests
from bs4 import BeautifulSoup
from stackapi import StackAPI, StackAPIError

class WebScraper:
    """A class to scrape the text content of a web page."""
    def __init__(self, url, session=None):
        self.url = url
        self.page_content = None
        self.soup = None
        self.session = session
        self.STACKOVERFLOW_SITE = StackAPI('stackoverflow', key=os.getenv("STACKOVERFLOW_API_KEY"))
        self._fetch_page()

    def fetch_stackoverflow_data(self, url):
        '''Fetches the content of a StackOverflow question and its answers.'''
        try:
            question_id = url.split('/')[-2]
            question = self.STACKOVERFLOW_SITE.fetch('questions/{ids}', ids=[question_id], filter='withbody')
            answers = self.STACKOVERFLOW_SITE.fetch('questions/{ids}/answers', ids=[question_id], filter='withbody')
            if question["items"]:
                combined_content = question["items"][0]["body"] + "\n"
                combined_content += "\n".join(answer["body"] for answer in answers["items"])
                return combined_content
            return None
        except StackAPIError as e:
            print(f"StackAPI error: {e}")
            return None
        except Exception as e:
            print(f"Error processing URL {url}: {e}")
            return None

    def _fetch_page(self):
        if "stackoverflow.com" in self.url:
            self.soup = None
            self.page_content = self.fetch_stackoverflow_data(self.url)
        else:
            try:
                response = self.session.get(self.url)
                response.encoding = 'utf-8'  # Ensure correct encoding
                self.page_content = response.content
                self.soup = BeautifulSoup(self.page_content, 'lxml')
            except requests.exceptions.RequestException as e:
                print(f"Error fetching the page: {e}")

    def get_text(self):
        """Get the text content of the web page."""
        if self.soup:

            text = self.soup.get_text()

            # Split text into lines and filter out empty ones
            lines = text.splitlines()
            lines = (line.strip() for line in lines if line.strip())

            # Split lines into phrases, strip, and filter out empty ones
            chunks = (phrase.strip() for line in lines for phrase in line.split("  ") if phrase.strip())

            # Join chunks with newline
            return '\n'.join(chunks)
        else:
            return self.page_content
