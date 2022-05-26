"""
Route module does all REST API operations on route endpoint
"""
import random
import urllib
from common import assert_status_code
from httpstatus import HTTP_OK, HTTP_NOT_FOUND, HTTP_CREATED, HTTP_BAD_REQUEST, HTTP_NO_CONTENT


class Route:
    """
    Route does all REST API operations on route endpoint
    """

    def __init__(self, config, session):
        self.config = config
        self.session = session
        self.url = config['api_url'] + '/v1/routes'


    def generate_route_name(self, number):
        """
        Returns name of test route based upon provided number
        """
        return f"testsuite-route{number}"


    def _create(self, success_expected, new_route=None):
        """
        Create new route to be used as test subject.
        If none provided generate route with random name.
        """

        if new_route is None:
            random_int = random.randint(0,99999)
            new_route = {
                "name" : self.generate_route_name(random_int),
                "displayName": f'TestSuite Route {random_int}',
                "path": f"/testsuite-path{random_int}",
                "pathType": "path",
                "attributes": [
                    {
                        "name": "HostHeader",
                        "value": f"testsuite{random_int}.example.com"
                    }
                ],
            }

        response = self.session.post(self.url, json=new_route)
        if success_expected:
            assert_status_code(response, HTTP_CREATED)

            # Check if just created route matches with what we requested
            created_route = response.json()
            self.assert_compare(created_route, new_route)
            return created_route

        assert_status_code(response, HTTP_BAD_REQUEST)
        return response.json()


    def create_positive(self, new_route=None):
        """
        Create new route, if none provided generate route with random data.
        """
        return self._create(True, new_route)


    def create_negative(self, route):
        """
        Attempt to create route which already exists, should fail
        """
        return self._create(False, route)


    def get_all(self):
        """
        Get all routes
        """
        response = self.session.get(self.url)
        assert_status_code(response, HTTP_OK)

        return response.json()


    def get_all_detailed(self):
        """
        Get all routes with full details
        """
        response = self.session.get(self.url + '?expand=true')
        assert_status_code(response, HTTP_OK)


    def get_positive(self, route):
        """
        Get existing route
        """
        response = self.session.get(self.url + '/' + urllib.parse.quote(route))
        assert_status_code(response, HTTP_OK)
        return response.json()


    def update_positive(self, route, updated_route):
        """
        Update existing route
        """
        response = self.session.post(self.url + '/' + urllib.parse.quote(route), json=updated_route)
        assert_status_code(response, HTTP_OK)
        return response.json()


    def _change_status(self, route_name, status, expect_success):
        """
        Update status of route
        """
        response = self.session.change_status(self.url + '/' + urllib.parse.quote(route_name), status)
        if expect_success:
            assert_status_code(response, HTTP_NO_CONTENT)
            assert response.content == b''
        else:
            assert_status_code(response, HTTP_BAD_REQUEST)


    def _delete(self, route_name, expected_success):
        """
        Delete existing route
        """
        route_url = self.url + '/' + urllib.parse.quote(route_name)
        response = self.session.delete(route_url)
        if expected_success:
            assert_status_code(response, HTTP_OK)
        else:
            assert_status_code(response, HTTP_NOT_FOUND)

        return response.json()


    def delete_positive(self, route_name):
        """
        Delete existing route
        """
        return self._delete(route_name, True)


    def delete_negative(self, route_name):
        """
        Attempt to delete non-existing route, should fail
        """
        return self._delete(route_name, False)


    def assert_compare(self, route_a, route_b):
        """
        Compares minimum required fields that can be set of two routes
        """
        assert route_a['name'] == route_b['name']
        assert (route_a['attributes'].sort(key=lambda x: x['name'])
            == route_b['attributes'].sort(key=lambda x: x['name']))
