"""
Organization module does all REST API operations on organization endpoint
"""
import random
import urllib
from common import assert_status_code
from httpstatus import HTTP_OK, HTTP_NOT_FOUND, HTTP_CREATED, HTTP_BAD_REQUEST


class Organization:
    """
    Organization does all REST API operations on organization endpoint
    """

    def __init__(self, config, session):
        self.config = config
        self.session = session
        self.url = config['api_url'] + '/v1/organizations'


    def generate_organization_name(self, number):
        """
        Returns name of test organization based upon provided number
        """
        return f"testsuite-organization{number}"


    def _create(self, success_expected, new_organization=None):
        """
        Create new organization to be used as test subject.
        If none provided generate organization with random name.
        """

        if new_organization is None:
            random_int = random.randint(0,99999)
            new_organization = {
                "name" : self.generate_organization_name(random_int),
                "displayName": f'TestSuite Organization {random_int}',
                "attributes" : [
                    {
                        "name" : f"name{random_int}",
                        "value" : f"value{random_int}",
                    }
                ],
            }

        response = self.session.post(self.url, json=new_organization)
        if success_expected:
            assert_status_code(response, HTTP_CREATED)

            # Check if just created organization matches with what we requested
            created_organization = response.json()
            self.assert_compare(created_organization, new_organization)

            return created_organization

        assert_status_code(response, HTTP_BAD_REQUEST)
        return response.json()


    def create_positive(self, new_organization=None):
        """
        Create new organization, if none provided generate organization with random data.
        """
        return self._create(True, new_organization)


    def create_negative(self, organization):
        """
        Attempt to create organization which already exists, should fail
        """
        return self._create(False, organization)


    def get_all(self):
        """
        Get all companies
        """
        response = self.session.get(self.url)
        assert_status_code(response, HTTP_OK)
        # TODO filtering on attributename, attributevalue, startKey

        return response.json()


    def get_all_detailed(self):
        """
        Get all companies with full details
        """
        response = self.session.get(self.url + '?expand=true')
        assert_status_code(response, HTTP_OK)
        # TODO filtering on attributename, attributevalue, startKey


    def get_positive(self, organization):
        """
        Get existing organization
        """
        response = self.session.get(self.url + '/' + urllib.parse.quote(organization))
        assert_status_code(response, HTTP_OK)
        return response.json()


    def update_positive(self, organization, updated_organization):
        """
        Update existing organization
        """
        response = self.session.post(self.url + '/' + urllib.parse.quote(organization), json=updated_organization)
        assert_status_code(response, HTTP_OK)
        return response.json()


    def _delete(self, organization_name, expected_success):
        """
        Delete existing organization
        """
        organization_url = self.url + '/' + urllib.parse.quote(organization_name)
        response = self.session.delete(organization_url)
        if expected_success:
            assert_status_code(response, HTTP_OK)
        else:
            assert_status_code(response, HTTP_NOT_FOUND)

        return response.json()


    def delete_positive(self, organization_name):
        """
        Delete existing organization
        """
        return self._delete(organization_name, True)


    def delete_negative(self, organization_name):
        """
        Attempt to delete non-existing organization, should fail
        """
        return self._delete(organization_name, False)


    def assert_compare(self, organization_a, organization_b):
        """
        Compares minimum required fields that can be set of two companies
        """
        assert organization_a['name'] == organization_b['name']
        assert (organization_a['attributes'].sort(key=lambda x: x['name'])
            == organization_b['attributes'].sort(key=lambda x: x['name']))
