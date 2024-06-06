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
from agents.title_gen import TitleAgent as TitleGen
from agents.log_filterer import LogFilterer

class GraphGen:
    """Class for the solution agent."""
    def __init__(self, model,tokenizer, endpoint_url: str, use_newspaper=False , tier: int = 1):
        load_dotenv()
        self.endpoint_url = endpoint_url
        self.tier = tier
        self.model = model
        self.tokenizer = tokenizer
        self.use_newspaper = use_newspaper
        self.load_agents()

    def load_agents(self):
        """Load the agents."""
        self.log_filterer = LogFilterer(self.endpoint_url, self.tier)
        self.title_generator = TitleGen(self.endpoint_url, self.tier)
        self.query_generator = QueryGen(self.endpoint_url, self.tier)
        self.search = GoogleCustomSearch(self.model, self.tokenizer, self.use_newspaper)
        self.ranker = Ranker(self.endpoint_url, self.tier)
        self.answer_generator = AnswerGenerator(self.endpoint_url, self.tier)


    def run(self, logs: str) -> Tuple[str, List[str]]:
        """Run the graph"""
        # Create function nodes
        filter_logs = Node(self.log_filterer.filter_relevant_logs, "Log Filterer", include_logs=True)
        logs = filter_logs.execute(logs=logs)

        gentitle = Node(self.title_generator.gen_title, "Title and Summary", include_logs=True)
        query_generator_node = Node(self.query_generator.gen_ques, "Query Generator", include_logs=True)
        search_node = Node(self.search.run_search, "Search")
        ranker_node = Node(self.ranker.rank, "Ranker", include_logs=True)
        answer_generator_node = Node(self.answer_generator.generate_answer, "Answer Generator", include_logs=True)

        # Connect nodes
        query_generator_node.add_child(search_node)  # Connect query generator to search
        search_node.add_child(ranker_node)  # Connect search to ranker
        ranker_node.add_child(answer_generator_node)  # Connect ranker to answer generator
        final_output, urls = query_generator_node.execute(logs=logs)
        header = gentitle.execute(logs=logs)
        output = {"title":header['title'], 
                  "Logsummary":header['logsummary'],
                  "predictedSolutions":final_output, 
                  "sources":urls,
                  "relevantLogs":logs}
        return output
    