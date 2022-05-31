"""
Listener module does all REST API operations on listener endpoint
"""
import random
import urllib
from common import assert_status_code
from httpstatus import HTTP_OK, HTTP_NOT_FOUND, HTTP_CREATED, HTTP_BAD_REQUEST, HTTP_NO_CONTENT


class Listener:
    """
    Listener does all REST API operations on listener endpoint
    """

    def __init__(self, config, session):
        self.config = config
        self.session = session
        self.url = config['api_url'] + '/v1/listeners'


    def generate_listener_name(self, number):
        """
        Returns name of test listener based upon provided number
        """
        return f"testsuite-listener{number}"


    def _create(self, success_expected, new_listener=None):
        """
        Create new listener to be used as test subject.
        If none provided generate listener with random name.
        """

        if new_listener is None:
            random_int = random.randint(0,99999)
            new_listener = {
                "name" : self.generate_listener_name(random_int),
                "displayName": f'TestSuite Listener {random_int}',
                "virtualHosts": [
                    f"testsuite{random_int}.example.com"
                ],
                "port": 80,
                "routeGroup": "123",
                "attributes": [
                    {
                        "name": "ServerName",
                        "value": f"testsuite{random_int} server"
                    }
                ],
            }

        response = self.session.post(self.url, json=new_listener)
        if success_expected:
            assert_status_code(response, HTTP_CREATED)

            # Check if just created listener matches with what we requested
            created_listener = response.json()
            self.assert_compare(created_listener, new_listener)
            return created_listener

        assert_status_code(response, HTTP_BAD_REQUEST)
        return response.json()


    def create_positive(self, new_listener=None):
        """
        Create new listener, if none provided generate listener with random data.
        """
        return self._create(True, new_listener)


    def create_negative(self, listener):
        """
        Attempt to create listener which already exists, should fail
        """
        return self._create(False, listener)


    def get_all(self):
        """
        Get all listeners
        """
        response = self.session.get(self.url)
        assert_status_code(response, HTTP_OK)
        return response.json()


    def get_positive(self, listener):
        """
        Get existing listener
        """
        response = self.session.get(self.url + '/' + urllib.parse.quote(listener))
        assert_status_code(response, HTTP_OK)
        return response.json()


    def update_positive(self, listener, updated_listener):
        """
        Update existing listener
        """
        response = self.session.post(self.url + '/' + urllib.parse.quote(listener), json=updated_listener)
        assert_status_code(response, HTTP_OK)
        return response.json()


    def _delete(self, listener_name, expected_success):
        """
        Delete existing listener
        """
        listener_url = self.url + '/' + urllib.parse.quote(listener_name)
        response = self.session.delete(listener_url)
        if expected_success:
            assert_status_code(response, HTTP_OK)
        else:
            assert_status_code(response, HTTP_NOT_FOUND)

        return response.json()


    def delete_positive(self, listener_name):
        """
        Delete existing listener
        """
        return self._delete(listener_name, True)


    def delete_negative(self, listener_name):
        """
        Attempt to delete non-existing listener, should fail
        """
        return self._delete(listener_name, False)


    def assert_compare(self, listener_a, listener_b):
        """
        Compares minimum required fields that can be set of two listeners
        """
        assert listener_a['name'] == listener_b['name']
        assert listener_a['virtualHosts'] == listener_b['virtualHosts']
        assert listener_a['port'] == listener_b['port']
        assert listener_a['routeGroup'] == listener_b['routeGroup']
        assert (listener_a['attributes'].sort(key=lambda x: x['name'])
            == listener_b['attributes'].sort(key=lambda x: x['name']))
