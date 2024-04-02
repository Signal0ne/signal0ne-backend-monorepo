import json
import os
from dotenv import load_dotenv
from typing import List, Tuple
from node.node import Node
from agents.answer import AnswerGenerator as AnswerGenerator
from websearch.search import GoogleCustomSearch
from agents.ranker import RankAgent as Ranker
from agents.query_gen import QueryAgent as QueryGen

class GraphGen:
    """Class for the solution agent."""
    def __init__(self, endpoint_url: str):
        load_dotenv()
        self.endpoint_url = endpoint_url
        self.load_agents()

    def load_agents(self):
        """Load the agents."""
        self.query_generator = QueryGen(self.endpoint_url)
        self.search = GoogleCustomSearch()
        self.ranker = Ranker(self.endpoint_url)
        self.answer_generator = AnswerGenerator(self.endpoint_url)

    def run(self, logs: str) -> Tuple[str, List[str]]:
        """Run the graph"""
        # Create function nodes
        query_generator_node = Node(self.query_generator.gen_ques, "Query Generator", include_logs=True)
        search_node = Node(self.search.run_search, "Search")
        ranker_node = Node(self.ranker.rank, "Ranker", include_logs=True)
        answer_generator_node = Node(self.answer_generator.generate_answer, "Answer Generator", include_logs=True)

        # Connect nodes
        query_generator_node.add_child(search_node)  # Connect query generator to search
        search_node.add_child(ranker_node)  # Connect search to ranker
        ranker_node.add_child(answer_generator_node)  # Connect ranker to answer generator

        # Execute the graph
        final_output, urls = query_generator_node.execute(logs=logs)
        output = [{"solution": final_output, "urls": urls}]
        return json.dumps(output)
    