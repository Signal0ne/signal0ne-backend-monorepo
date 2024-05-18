
def parse_json(text_string):
    # Trim result to get only the json output
    firstPos = text_string.index("{")
    lastPos = text_string[::-1].index("}")
    return text_string[firstPos:len(text_string)-lastPos]