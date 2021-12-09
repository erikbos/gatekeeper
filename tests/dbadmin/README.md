# managementserver tests

This directory contains test suite to test developer endpoints.

## Environment configuration

Environment variable required to be set before running tests:

* API_URL, base URL (including organization path, if set)
* API_USERNAME, username used for HTTP Basic authentication
* API_PASSWORD, password used for HTTP Basic authentication

## Example

    # Set envs

    cd tests/managementserver

    export API_URL=<https://managementserver/v1/organization/default>
    export API_USERNAME=admin
    export API_PASSWORD=passwd

    # Run all developer related tests
    pytest -v test_developer

    # Run five tests in parallel
    pytest -n 5 test_developer

    # Run one test named 'test_crud_lifecycle_10_developers'
    pytest -k test_get_all_developers_list
