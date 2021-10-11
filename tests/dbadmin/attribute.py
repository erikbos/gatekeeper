from common import assert_valid_schema, assert_status_code, assert_content_type_json, load_json_schema
from httpstatus import HTTP_OK, HTTP_NOT_FOUND


class Attribute:
    """
    Attribute does all REST API functions on developer endpoint
    """
    def __init__(self, config, session, attribute_url):
        self.config = config
        self.session = session
        self.attribute_url = attribute_url
        self.schemas = {
            'attribute': load_json_schema('attribute.json'),
            'attributes': load_json_schema('attributes.json'),
            'error': load_json_schema('error.json'),
        }


    def post(self, attributes):
        """
        Update all attributes
        """
        response = self.session.post(self.attribute_url, json=attributes)
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)
        assert_valid_schema(response.json(), self.schemas['attributes'])


    def get_all(self):
        """
        Retrieve all attributes
        """
        response = self.session.get(self.attribute_url)
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)

        retrieved_attributes = response.json()
        assert_valid_schema(retrieved_attributes, self.schemas['attributes'])

        return retrieved_attributes


    def get_existing(self, attribute_name):
        """
        Retrieve an existing attribute
        """
        response = self.session.get(self.attribute_url + '/' + attribute_name)
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)

        retrieved_attribute = response.json()
        assert_valid_schema(retrieved_attribute, self.schemas['attribute'])

        return retrieved_attribute


    def get_non_existing(self, attribute_name):
        """
        Attempt to retrieve a non-existing attribute
        """
        response = self.session.get(self.attribute_url + '/' + attribute_name)
        assert_status_code(response, HTTP_NOT_FOUND)
        assert_content_type_json(response)


    def delete(self, attribute_name):
        """
        Delete an existing attribute
        """
        response = self.session.delete(self.attribute_url + '/' + attribute_name)
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)

        retrieved_attribute = response.json()
        assert_valid_schema(retrieved_attribute, self.schemas['attribute'])

        return retrieved_attribute


    def delete_non_existing(self, attribute_name):
        """
        Attempt to delete a non-existing attribute
        """
        response = self.session.delete(self.attribute_url + '/' + attribute_name)
        assert_status_code(response, HTTP_NOT_FOUND)
        assert_content_type_json(response)


    def compare_sets(self, attribute_response_a, attribute_response_b):
        """
        Compare two attribute response sets
        """
        assert [i for i in attribute_response_a['attribute'] if i not in attribute_response_b['attribute']] == []


def run_attribute_tests(config, session, attribute_url):
    """
    Test all operations that can be done on an attribute endpoints
    """

    attributeAPI = Attribute(config, session, attribute_url)

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
