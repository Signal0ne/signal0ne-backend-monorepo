"""This module contains the Node class."""
class Node:
    def __init__(self, func, name):
        self.func = func
        self.name = name
        self.children = []

    def add_child(self, child):
        """Add a child node to the current node."""
        self.children.append(child)

    def execute(self, input_data):
        """Execute the function of the node and pass the output to the children."""
        print("Executing function:", self.name)
        output = self.func(input_data)
        for child in self.children:
            output = child.execute(output)
        return output
    