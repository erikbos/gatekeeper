import os
import jsonref
import jsonschema


HTTP_OK = 200
HTTP_CREATED = 201
HTTP_BAD_REQUEST = 400
HTTP_AUTHORIZATION_REQUIRED = 401
HTTP_NOT_FOUND = 404
HTTP_BAD_CONTENT = 406


# Required for each request
default_headers = {
    'accept': 'application/json',
      'user-agent': 'Gatekeeper testsuite'
    }


def get_config():
    """Returns endpoint configuration"""

    return {
        "api_url": os.environ['API_URL'],
        "api_username": os.environ['API_USERNAME'],
        "api_password": os.environ['API_PASSWORD'],
        "request_headers": default_headers,
    }


def assert_status_code(response, status_code):
    """Checks response status code"""
    assert response.status_code == status_code


def assert_content_type_json(response):
    """Checks content-type is json"""
    assert response.headers['content-type'].startswith('application/json')


def assert_valid_schema(data, schema_file):
    """
    Checks whether the given data matches the schema
    """
    schema = _load_json_schema(schema_file)
    return jsonschema.validate(data, schema)


def _load_json_schema(filename):
    """
    Loads the given schema file
    """
    relative_path = os.path.join('schemas', filename)
    absolute_path = os.path.join(os.path.dirname(__file__), relative_path)

    base_path = os.path.dirname(absolute_path)
    base_uri = 'file://{}/'.format(base_path)

    with open(absolute_path) as schema_file:
        return jsonref.loads(schema_file.read(), base_uri=base_uri, jsonschema=True)
