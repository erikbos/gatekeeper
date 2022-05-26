"""
Attribute module does all REST API operations on an attribute endpoint
"""
import urllib
from common import assert_status_code
from httpstatus import HTTP_OK, HTTP_NOT_FOUND


class Attribute:
    """
    Attribute does all REST API functions on developer endpoint
    """
    def __init__(self, config, session, attribute_url):
        self.config = config
        self.session = session
        self.url = attribute_url


    def post(self, attributes):
        """
        Update all attributes
        """
        response = self.session.post(self.url, json=attributes)
        assert_status_code(response, HTTP_OK)


    def get_all(self):
        """
        Retrieve all attributes
        """
        response = self.session.get(self.url)
        assert_status_code(response, HTTP_OK)
        return response.json()


    def get_positive(self, attribute_name):
        """
        Retrieve an existing attribute
        """
        response = self.session.get(self.url + '/' + urllib.parse.quote(attribute_name))
        assert_status_code(response, HTTP_OK)
        return response.json()


    def get_negative(self, attribute_name):
        """
        Attempt to retrieve a non-existing attribute
        """
        response = self.session.get(self.url + '/' + urllib.parse.quote(attribute_name))
        assert_status_code(response, HTTP_NOT_FOUND)


    def delete_positive(self, attribute_name):
        """
        Delete an existing attribute
        """
        response = self.session.delete(self.url + '/' + urllib.parse.quote(attribute_name))
        assert_status_code(response, HTTP_OK)
        return response.json()


    def delete_negative(self, attribute_name):
        """
        Attempt to delete a non-existing attribute
        """
        response = self.session.delete(self.url + '/' + urllib.parse.quote(attribute_name))
        assert_status_code(response, HTTP_NOT_FOUND)


    def compare_sets(self, attribute_response_a, attribute_response_b):
        """
        Compare two attribute response sets
        """
        assert [i for i in attribute_response_a['attribute'] \
            if i not in attribute_response_b['attribute']] == []


def run_attribute_tests(config, session, attribute_url):
    """
    Test all operations that can be done on an attribute endpoints
    """
    attribute_api = Attribute(config, session, attribute_url)

    # Update attributes to known set
    attribute_status = {
        "name" : "Status",
        "value" : ""
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
    attribute_api.post(new_attributes)

    # Did all attributes get stored?
    retrieved_attributes = attribute_api.get_all()
    attribute_api.compare_sets(retrieved_attributes, new_attributes)

    # Retrieve single attribute
    assert attribute_api.get_positive(attribute_status['name']) == attribute_status
    assert attribute_api.get_positive(attribute_shoesize['name']) == attribute_shoesize

    # Delete one attribute and attempt to fetch it again
    attribute_api.delete_positive(attribute_status['name'])
    attribute_api.get_negative(attribute_status['name'])

    # Attempt to delete attribute that just got deleted
    attribute_api.delete_negative(attribute_status['name'])
