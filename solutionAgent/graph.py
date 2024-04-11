import json
import os
import re
from dotenv import load_dotenv
from typing import List, Tuple
from node.node import Node
from agents.answer import AnswerGenerator as AnswerGenerator
from websearch.search import GoogleCustomSearch
from agents.ranker import RankAgent as Ranker
from agents.query_gen import QueryAgent as QueryGen
from langchain_community.llms import HuggingFaceEndpoint

class GraphGen:
    """Class for the solution agent."""
    def __init__(self, endpoint_url: str):
        load_dotenv()
        self.endpoint_url = endpoint_url
        self.llm = HuggingFaceEndpoint(
            endpoint_url=self.endpoint_url,
            task="text-generation",
            max_new_tokens=512,
            top_k=50,
            temperature=0.4,
            repetition_penalty=1.1,
        )
        self.load_agents()

    def load_agents(self):
        """Load the agents."""
        self.query_generator = QueryGen(self.endpoint_url)
        self.search = GoogleCustomSearch()
        self.ranker = Ranker(self.endpoint_url)
        self.answer_generator = AnswerGenerator(self.endpoint_url)

    def genTitleSummary(self, logs: str) -> Tuple[str, str]:
        """Generate title and summary of the error logs."""
        ans = self.llm(f"""Act like a software debugger and generate a title for these
                         error logs and give a single paragraph summary of the error logs in technical detail in a json format.
                         logs: {logs})
                         Output json format is {{\"title\": \"your title\", \"logsummary\": \"your summary\"}}
                         json: """)
        json_str = ans[ans.find('{'):ans.rfind('}') + 1]


        if json_str:
            result = json.loads(json_str)
            title = result["title"]
            summary = result["logsummary"]
        else:
            title = ""
            summary = ""
        return title, summary

    def run(self, logs: str) -> Tuple[str, List[str]]:
        """Run the graph"""
        # Create function nodes
        gentitle = Node(self.genTitleSummary, "Title and Summary", include_logs=True)
        query_generator_node = Node(self.query_generator.gen_ques, "Query Generator", include_logs=True)
        search_node = Node(self.search.run_search, "Search")
        ranker_node = Node(self.ranker.rank, "Ranker", include_logs=True)
        answer_generator_node = Node(self.answer_generator.generate_answer, "Answer Generator", include_logs=True)

        # Connect nodes
        query_generator_node.add_child(search_node)  # Connect query generator to search
        search_node.add_child(ranker_node)  # Connect search to ranker
        ranker_node.add_child(answer_generator_node)  # Connect ranker to answer generator
        final_output, urls = query_generator_node.execute(logs=logs)
        title, summary = gentitle.execute(logs=logs)
        output = {"title":title, "Logsummary":summary,"predictedSolutions":final_output, "sources":urls}
        return output
    