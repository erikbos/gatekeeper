"""
Provides common functions
"""
import os
import jsonref
import jsonschema


def get_config():
    """
    Returns endpoint configuration
    """
    required_envs = ['API_URL', 'API_USERNAME', 'API_PASSWORD']
    for var in required_envs:
        if var not in os.environ:
            raise EnvironmentError(f'Please set {var} as configuration parameter.')

    return {
        'api_url': os.environ['API_URL'],
        'api_username': os.environ['API_USERNAME'],
        'api_password': os.environ['API_PASSWORD'],
        'entity_count': 3,
    }


def assert_status_code(response, status_code):
    """
    Checks response status code
    """
    assert response.status_code == status_code


def assert_content_type_json(response):
    """
    Checks content-type is json
    """
    assert response.headers['content-type'].startswith('application/json')


def assert_valid_schema(data, schema):
    """
    Checks whether the given data matches the schema
    """
    return jsonschema.validate(data, schema)


def assert_valid_schema_error(data):
    """
    Checks whether the given data matches the schema
    """
    return jsonschema.validate(data, load_json_schema('error.json'))


def load_json_schema(filename):
    """
    Loads the given schema file
    """
    relative_path = os.path.join('schemas', filename)
    absolute_path = os.path.join(os.path.dirname(__file__), relative_path)

    base_path = os.path.dirname(absolute_path)
    base_uri = f'file://{base_path}/'

    with open(absolute_path, encoding='utf-8') as schema_file:
        return jsonref.loads(schema_file.read(), base_uri=base_uri, jsonschema=True)
