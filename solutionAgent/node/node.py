class Node:
    """A class to represent a node in a tree structure. Each node is a function"""
    def __init__(self, func, name, include_logs=False):
        self.func = func
        self.name = name
        self.children = []
        self.include_logs = include_logs

    def add_child(self, child):
        """Add a child node to the current node."""
        self.children.append(child)

    def execute(self, *input_data, **kwargs):
        """Execute the function of the node and pass the output to the children."""
        print("Executing function:", self.name)
                
        if not self.include_logs:
            filtered_kwargs = {k: v for k, v in kwargs.items() if k != 'logs'}
        else:
            filtered_kwargs = kwargs

        # Pass only required number of inputs
        if input_data and isinstance(input_data[0], tuple):
            output = self.func(*input_data[0], **filtered_kwargs)
        else:
            output = self.func(*input_data, **filtered_kwargs)
        
        # Execute child nodes
        for child in self.children:
            output = child.execute(output, **kwargs)
        
        return output
