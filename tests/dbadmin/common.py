"""
Provides common functions
"""
import os
import requests
import urllib
from openapi_core import create_spec
from openapi_spec_validator.schemas import read_yaml_file
from openapi_core.contrib.requests import RequestsOpenAPIRequest, RequestsOpenAPIResponse
from openapi_core.validation.request.validators import RequestValidator
from openapi_core.validation.response.validators import ResponseValidator
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


def get_http_session(config):
    """
    Returns HTTP session for all API requests
    """
    session = requests.Session()
    session.auth = (config['api_username'], config['api_password'])
    session.headers = {
        'accept': 'application/json',
        'user-agent': 'Gatekeeper testsuite'
        }
    return session


def assert_status_code(response, status_code):
    """
    Checks response status code
    """
    assert response.status_code == status_code


def assert_status_codes(response, status_codes):
    """
    Checks one or more response status codes
    """
    assert [ True for code in status_codes if code == response.status_code]


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

class API:
    def __init__(self, config, filename):
        self.config = config
        self.session = requests.Session()
        self.session.auth = (config['api_username'], config['api_password'])
        self.headers = {
            'accept': 'application/json',
            'user-agent': 'Gatekeeper testsuite'
            }

        spec = create_spec(read_yaml_file(filename))
        self.request_validator = RequestValidator(spec)
        self.response_validator = ResponseValidator(spec)


    def _request(self, method, url, **kwargs):
        request = requests.Request(method, url, **kwargs)
        result = self.request_validator.validate(RequestsOpenAPIRequest(request))
        result.raise_for_errors()

        response = self.session.send(self.session.prepare_request(request))
        responseOpenAPI = RequestsOpenAPIResponse(response)

        # Workaround for https://github.com/p1c2u/openapi-core/issues/378
        # We override received content-type with the one in our OpenAPI specification
        if responseOpenAPI.mimetype == 'application/json; charset=utf-8':
            responseOpenAPI.mimetype = 'application/json'

        result = self.response_validator.validate(RequestsOpenAPIRequest(request), responseOpenAPI)

        result.raise_for_errors()

        return response

    def get(self, url, **kwargs):
        return self._request("GET", url, **kwargs)

    def post(self, url, **kwargs):
        return self._request("POST", url, **kwargs)

    def put(self, url, **kwargs):
        return self._request("PUT", url, **kwargs)

    def delete(self, url):
        return self._request("DELETE", url)

    def change_status(self, url, status):
        headers = self.session.headers
        headers['content-type'] = 'application/octet-stream'
        return self.post(url + '?action=' + status, headers=headers)
