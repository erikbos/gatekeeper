# dbadmin tests

This directory contains test suite to test developer endpoints.

## Environment configuration

Environment variable required to be set before running tests:

* API_URL, base URL (including organization path, if set)
* API_USER, username used for HTTP Basic authentication
* API_PASSWORD, password used for HTTP Basic authentication

## Example

    # Set envs
    export API_URL=<https://dbadmin/v1/organization/default>
    export API_USER=admin
    export API_PASSWORD=passwd

    # Run all developer related tests
    pytest tests/test_developer.py

    # Run five tests in parallel
    pytest -n 5 test_developer.py

    # Run one test named 'test_crud_lifecycle_10_developers'
    pytest -k test_get_all_developers_list test_developer.py
