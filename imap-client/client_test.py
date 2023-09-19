import unittest

from unittest.mock import patch
from client import get_group_settings_to_process, should_process_email, valid_from_email, valid_subject, valid_mime_type


class TestGetGroupSettingsToProcess(unittest.TestCase):

    def test_empty_lists(self):
        self.assertEqual(get_group_settings_to_process([], []), [])

    def test_no_matching_emails(self):
        group_settings = [{"emailToRead": "test1@example.com"},
                          {"emailToRead": "test2@example.com"}]
        emailSettings = [{"username": "no_match@example.com"}]
        self.assertEqual(get_group_settings_to_process(
            group_settings, emailSettings), [])

    def test_some_matching_emails(self):
        group_settings = [{"emailToRead": "test1@example.com"},
                          {"emailToRead": "test2@example.com"}]
        emailSettings = [{"username": "test1@example.com"},
                         {"username": "no_match@example.com"}]
        self.assertEqual(get_group_settings_to_process(
            group_settings, emailSettings), [{"emailToRead": "test1@example.com"}])

    def test_all_matching_emails(self):
        group_settings = [{"emailToRead": "test1@example.com"},
                          {"emailToRead": "test2@example.com"}]
        emailSettings = [{"username": "test1@example.com"},
                         {"username": "test2@example.com"}]
        self.assertEqual(get_group_settings_to_process(
            group_settings, emailSettings), group_settings)

    def test_extra_emailSettings(self):
        group_settings = [{"emailToRead": "test1@example.com"}]
        emailSettings = [{"username": "test1@example.com"},
                         {"username": "extra@example.com"}]
        self.assertEqual(get_group_settings_to_process(
            group_settings, emailSettings), group_settings)

    def test_extra_group_settings(self):
        group_settings = [{"emailToRead": "test1@example.com"},
                          {"emailToRead": "extra@example.com"}]
        emailSettings = [{"username": "test1@example.com"}]
        self.assertEqual(get_group_settings_to_process(
            group_settings, emailSettings), [{"emailToRead": "test1@example.com"}])


class TestValidFromEmail(unittest.TestCase):

    def test_empty_whitelist(self):
        self.assertTrue(valid_from_email("test@example.com", []))

    def test_email_not_in_whitelist(self):
        whitelist = [{"email": "allowed@example.com"}]
        self.assertFalse(valid_from_email("test@example.com", whitelist))

    def test_email_in_whitelist(self):
        whitelist = [{"email": "allowed@example.com"}]
        self.assertTrue(valid_from_email("allowed@example.com", whitelist))

    def test_multiple_emails_in_whitelist(self):
        whitelist = [{"email": "allowed1@example.com"},
                     {"email": "allowed2@example.com"}]
        self.assertTrue(valid_from_email("allowed1@example.com", whitelist))
        self.assertTrue(valid_from_email("allowed2@example.com", whitelist))

    def test_multiple_emails_not_in_whitelist(self):
        whitelist = [{"email": "allowed1@example.com"},
                     {"email": "allowed2@example.com"}]
        self.assertFalse(valid_from_email(
            "not_allowed@example.com", whitelist))

    def test_email_case_sensitivity(self):
        whitelist = [{"email": "Allowed@example.com"}]
        self.assertFalse(valid_from_email("allowed@example.com", whitelist))


class TestValidSubject(unittest.TestCase):

    def test_empty_regex_list(self):
        self.assertTrue(valid_subject("Some Subject", []))

    def test_no_match(self):
        subject_line_regexes = [{"regex": "Urgent"}]
        self.assertFalse(valid_subject("Some Subject", subject_line_regexes))

    def test_single_match(self):
        subject_line_regexes = [{"regex": "Urgent"}]
        self.assertTrue(valid_subject(
            "Urgent: Important email", subject_line_regexes))

    def test_multiple_regexes_no_match(self):
        subject_line_regexes = [{"regex": "Urgent"}, {"regex": "Important"}]
        self.assertFalse(valid_subject("Random Subject", subject_line_regexes))

    def test_multiple_regexes_single_match(self):
        subject_line_regexes = [{"regex": "Urgent"}, {"regex": "Important"}]
        self.assertTrue(valid_subject(
            "Important: Please read", subject_line_regexes))

    def test_multiple_regexes_multiple_match(self):
        subject_line_regexes = [{"regex": "Urgent"}, {"regex": "Important"}]
        self.assertTrue(valid_subject(
            "Urgent and Important: Please read", subject_line_regexes))

    def test_case_sensitive_no_match(self):
        subject_line_regexes = [{"regex": "URGENT"}]
        self.assertFalse(valid_subject(
            "Urgent: Please read", subject_line_regexes))

    def test_case_sensitive_match(self):
        subject_line_regexes = [{"regex": "URGENT"}]
        self.assertTrue(valid_subject(
            "URGENT: Please read", subject_line_regexes))

    def test_regex_special_characters(self):
        subject_line_regexes = [{"regex": r"\d{4}"}]  # Matches four digits
        self.assertTrue(valid_subject("Code: 1234", subject_line_regexes))
        self.assertFalse(valid_subject("Code: ABCD", subject_line_regexes))


class TestValidMimeType(unittest.TestCase):

    def test_valid_image_mime_types(self):
        self.assertTrue(valid_mime_type("image/png"))
        self.assertTrue(valid_mime_type("image/jpeg"))
        self.assertTrue(valid_mime_type("image/gif"))

    def test_invalid_mime_types(self):
        self.assertFalse(valid_mime_type("text/plain"))
        self.assertFalse(valid_mime_type("audio/mp3"))
        self.assertFalse(valid_mime_type("application/json"))

    def test_case_sensitivity(self):
        self.assertFalse(valid_mime_type("IMAGE/PNG"))
        self.assertFalse(valid_mime_type("ImAgE/jPeG"))

    def test_empty_string(self):
        self.assertFalse(valid_mime_type(""))

    def test_whitespace(self):
        self.assertFalse(valid_mime_type(" "))

    def test_none_input(self):
        self.assertFalse(valid_mime_type(None))


class TestShouldProcessEmail(unittest.TestCase):

    @patch('client.valid_from_email', return_value=True)
    @patch('client.valid_subject', return_value=True)
    def test_both_valid(self, mock_valid_subject, mock_valid_email):
        group_setting = {"emailWhiteList": "anything",
                         "subjectLineRegexes": "anything"}
        self.assertTrue(should_process_email(
            "Valid Subject", "valid@example.com", group_setting))

    @patch('client.valid_from_email', return_value=True)
    @patch('client.valid_subject', return_value=False)
    def test_invalid_subject(self, mock_valid_subject, mock_valid_email):
        group_setting = {"emailWhiteList": "anything",
                         "subjectLineRegexes": "anything"}
        self.assertFalse(should_process_email(
            "Invalid Subject", "valid@example.com", group_setting))

    @patch('client.valid_from_email', return_value=False)
    @patch('client.valid_subject', return_value=True)
    def test_invalid_email(self, mock_valid_subject, mock_valid_email):
        group_setting = {"emailWhiteList": "anything",
                         "subjectLineRegexes": "anything"}
        self.assertFalse(should_process_email(
            "Valid Subject", "invalid@example.com", group_setting))

    @patch('client.valid_from_email', return_value=False)
    @patch('client.valid_subject', return_value=False)
    def test_both_invalid(self, mock_valid_subject, mock_valid_email):
        group_setting = {"emailWhiteList": "anything",
                         "subjectLineRegexes": "anything"}
        self.assertFalse(should_process_email(
            "Invalid Subject", "invalid@example.com", group_setting))


if __name__ == '__main__':
    unittest.main()
