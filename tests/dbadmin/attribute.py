import requests
from common import assert_valid_schema, assert_status_code, assert_content_type_json, load_json_schema
from httpstatus import HTTP_OK, HTTP_NOT_FOUND


class Attribute:
    """
    Attribute does all REST API functions on developer endpoint
    """
    def __init__(self, config, attribute_url):
        self.config = config
        self.attribute_url = attribute_url
        self.schema_attribute = load_json_schema('attribute.json')
        self.schema_attributes = load_json_schema('attributes.json')
        self.schema_error = load_json_schema('error.json')

    def _get_auth(self):
        """
        Returns HTTP auth parameters
        """
        if 'api_username' in self.config and 'api_password' in self.config:
            return (self.config['api_username'], self.config['api_password'])
        else:
            return None


    def post(self, attributes):
        """
        Update all attributes
        """
        response = requests.post(self.attribute_url, auth=self._get_auth(), headers=self.config['request_headers'], json=attributes)
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)
        assert_valid_schema(response.json(), self.schema_attributes)


    def get_all(self):
        """
        Retrieve all attributes
        """
        response = requests.get(self.attribute_url, auth=self._get_auth(), headers=self.config['request_headers'])
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)

        retrieved_attributes = response.json()
        assert_valid_schema(retrieved_attributes, self.schema_attributes)

        return retrieved_attributes


    def get_existing(self, attribute_name):
        """
        Retrieve an existing attribute
        """
        response = requests.get(self.attribute_url + '/' + attribute_name, auth=self._get_auth(), headers=self.config['request_headers'])
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)

        retrieved_attribute = response.json()
        assert_valid_schema(retrieved_attribute, self.schema_attribute)

        return retrieved_attribute


    def get_non_existing(self, attribute_name):
        """
        Attempt to retrieve a non-existing attribute
        """
        response = requests.get(self.attribute_url + '/' + attribute_name, auth=self._get_auth(), headers=self.config['request_headers'])
        assert_status_code(response, HTTP_NOT_FOUND)
        assert_content_type_json(response)


    def delete(self, attribute_name):
        """
        Delete an existing attribute
        """
        response = requests.delete(self.attribute_url + '/' + attribute_name, auth=self._get_auth(), headers=self.config['request_headers'])
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)

        retrieved_attribute = response.json()
        assert_valid_schema(retrieved_attribute, self.schema_attribute)

        return retrieved_attribute


    def delete_non_existing(self, attribute_name):
        """
        Attempt to delete a non-existing attribute
        """
        response = requests.delete(self.attribute_url + '/' + attribute_name, auth=self._get_auth(), headers=self.config['request_headers'])
        assert_status_code(response, HTTP_NOT_FOUND)
        assert_content_type_json(response)


    def compare_sets(self, attribute_response_a, attribute_response_b):
        """
        Compare two attribute response sets
        """
        assert [i for i in attribute_response_a['attribute'] if i not in attribute_response_b['attribute']] == []


def run_attribute_tests(config, attribute_url):
    """
    Test all operations that can be done on an attribute endpoints
    """

    attributeAPI = Attribute(config, attribute_url)

    # Update attributes to known set
    attribute_status = {
        "name" : "Status"
    }
    attribute_shoesize = {
        "name" : "Shoesize",
        "value" : "42"
    }
    new_attributes = {
        "attribute": [
            attribute_status,
            attribute_shoesize
        ]
    }
    attributeAPI.post(new_attributes)

    # Did all attributes get stored?
    retrieved_attributes = attributeAPI.get_all()
    attributeAPI.compare_sets(retrieved_attributes, new_attributes)

    # Retrieve single attributes
    assert attributeAPI.get_existing(attribute_status['name']) == attribute_status
    assert attributeAPI.get_existing(attribute_shoesize['name']) == attribute_shoesize

    # Delete one attribute
    attributeAPI.delete(attribute_status['name'])
    attributeAPI.get_non_existing(attribute_status['name'])

    # Attempt to delete attribute that just got deleted
    attributeAPI.delete_non_existing(attribute_status['name'])
