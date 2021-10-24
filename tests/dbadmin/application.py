"""
Application module does all REST API operations on application endpoint
"""
import random
import urllib
from common import assert_valid_schema, assert_status_code, assert_content_type_json, \
    load_json_schema
from httpstatus import HTTP_OK, HTTP_NOT_FOUND, HTTP_CREATED, HTTP_BAD_REQUEST


class Application:
    """
    Application does all REST API functions on application endpoint
    """
    def __init__(self, config, session, developer_email):
        self.config = config
        self.session = session
        if developer_email is not None:
            self.application_url = config['api_url'] + '/developers/' + developer_email + '/apps'
        self.schemas = {
            'application': load_json_schema('application.json'),
            'applications': load_json_schema('applications.json'),
            'applications-global-list': load_json_schema('applications-global-list.json'),
            'error': load_json_schema('error.json'),
        }


    def generate_app_name(self, number):
        """
        Returns generated test application name
        """
        return f"testsuite-app{number}"


    def create_new(self, new_application=None):
        """
        Create new application to be used as test subject.
        If no app data provided generate application with random data
        """
        if new_application is None:
            random_int = random.randint(0,99999)
            new_application = {
                "name" : self.generate_app_name(random_int),
                "attributes" : [
                    {
                        "name" : f"name{random_int}",
                        "value" : f"value{random_int}",
                    }
                ],
            }

        response = self.session.post(self.application_url, json=new_application)
        assert_status_code(response, HTTP_CREATED)
        assert_content_type_json(response)

        # Check if just created application matches with what we requested
        created_application = response.json()
        assert_valid_schema(created_application, self.schemas['application'])
        self.assert_compare(created_application, new_application)

        return created_application


    def create_existing(self, application):
        """
        Attempt to create new application which already exists, should fail
        """
        response = self.session.post(self.application_url, json=application)
        assert_status_code(response, HTTP_BAD_REQUEST)
        assert_content_type_json(response)
        assert_valid_schema(response.json(), self.schemas['error'])


    def get_all_global(self):
        """
        Get all apps of all developers (uuid list)
        """
        response = self.session.get(self.config['api_url'] + '/apps')
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)
        assert_valid_schema(response.json(), self.schemas['applications-global-list'])

        return response.json()


    def get_all(self):
        """
        Get all applications of one developer
        """
        response = self.session.get(self.application_url)
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)
        assert_valid_schema(response.json(), self.schemas['applications'])

        return response.json()


    def get_existing(self, app_name):
        """
        Get existing application
        """
        developer_url = self.application_url + '/' + urllib.parse.quote(app_name)
        response = self.session.get(developer_url)
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)
        retrieved_developer = response.json()
        assert_valid_schema(retrieved_developer, self.schemas['application'])

        return retrieved_developer


    def update_existing(self, application):
        """
        Update existing application
        """
        application_url = self.application_url + '/' + urllib.parse.quote(application['name'])
        response = self.session.post(application_url, json=application)
        # print(app_name)
        # print(application_url)
        # print(application)
        # print(response.text)
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)

        updated_application = response.json()
        assert_valid_schema(updated_application, self.schemas['application'])

        return updated_application


    def delete_existing(self, app_name):
        """
        Delete existing application
        """
        application_url = self.application_url + '/' + urllib.parse.quote(app_name)
        response = self.session.delete(application_url)
        assert_status_code(response, HTTP_OK)
        assert_content_type_json(response)
        updated_application = response.json()
        assert_valid_schema(updated_application, self.schemas['application'])

        return updated_application


    def delete_nonexisting(self, app_name):
        """
        Delete non-existing application, which should fail
        """
        application_url = self.application_url + '/' + urllib.parse.quote(app_name)
        response = self.session.delete(application_url)
        assert_status_code(response, HTTP_NOT_FOUND)
        assert_content_type_json(response)


    def delete_all_test_developer(self):
        """
        Delete all application created by test suite
        """
        for i in range(self.config['entity_count']):
            app_name = self.generate_app_name(i)
            application_url = self.application_url + '/' + urllib.parse.quote(app_name)
            self.session.delete(application_url)


    def assert_compare(self, application_a, application_b):
        """
        Compares minimum required fields that can be set of two applications
        """
        assert application_a['name'] == application_b['name']
        assert application_a['attributes'] == application_b['attributes']
