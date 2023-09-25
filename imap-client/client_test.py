import datetime
import tempfile
import unittest

from unittest.mock import Mock, patch
from client import get_attachments, get_formatted_message_data, get_group_settings_to_process, getFormattedDate, getFromData, should_process_email, valid_from_email, valid_subject, valid_mime_type


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


class TestGetFormattedMessageData(unittest.TestCase):

    @patch('email.message_from_bytes')
    @patch('client.should_process_email')
    @patch('client.get_attachments')
    def test_process_email(self, mock_get_attachments, mock_should_process_email, mock_message_from_bytes):
        # Mock email message
        mock_message = Mock()
        mock_message.get.side_effect = lambda x: {
            "From": "John Doe <john@example.com>",
            "To": "to@example.com",
            "Subject": "Test Subject",
            "Date": "Mon, 20 Sep 2021 10:10:10 +0000"
        }.get(x, "")
        mock_message_from_bytes.return_value = mock_message

        # Mock should_process_email
        mock_should_process_email.return_value = True

        # Mock get_attachments
        mock_get_attachments.return_value = [{}, {}]

        # Set up function arguments
        data = {b"RFC822": b"email data"}
        group_settings_to_process = [
            {"emailToRead": "to@example.com", "id": 1},
            {"emailToRead": "another_to@example.com", "id": 2}
        ]

        result = get_formatted_message_data(data, group_settings_to_process)

        # TODO: Check date correctly
        self.assertEqual(result["date"], "2021-09-20T10:10:10.000000Z")
        self.assertEqual(result["subject"], "Test Subject")
        self.assertEqual(result["to"], "to@example.com")
        self.assertEqual(result["fromName"], "John Doe ")
        self.assertEqual(result["fromEmail"], "john@example.com")
        self.assertEqual(result["groupSettingsIds"], [1])
        self.assertEqual(result["attachments"], [{}, {}])

    @patch('email.message_from_bytes')
    @patch('client.should_process_email')
    @patch('datetime.datetime')
    @patch('client.get_attachments')
    def test_no_process_email(self, mock_get_attachments, mock_datetime, mock_should_process_email, mock_message_from_bytes):
        # Mock email message
        mock_message = Mock()
        mock_message.get.side_effect = lambda x: {
            "From": "John Doe <john@example.com>",
            "To": "to@example.com",
            "Subject": "Test Subject",
            "Date": "Mon, 20 Sep 2021 10:10:10 +0000"
        }.get(x, "")
        mock_message_from_bytes.return_value = mock_message

        # Mock should_process_email
        mock_should_process_email.return_value = False

        data = {b"RFC822": b"email data"}
        group_settings_to_process = [
            {"emailToRead": "to@example.com", "id": 1},
            {"emailToRead": "another_to@example.com", "id": 2}
        ]

        result = get_formatted_message_data(data, group_settings_to_process)

        self.assertEqual(result, {})


class TestGetFromData(unittest.TestCase):

    # def setUp(self):
    # logging.disable(logging.CRITICAL)  # to disable logging for tests

    def test_split_length_two(self):
        message_data = {"From": "John Doe <john.doe@example.com>"}
        result = getFromData(message_data)
        self.assertEqual(result, {
            "fromName": "John Doe ",
            "fromEmail": "john.doe@example.com"
        })

    def test_split_length_one(self):
        message_data = {"From": "john.doe@example.com"}
        result = getFromData(message_data)
        self.assertEqual(result, {
            "fromName": None,
            "fromEmail": "john.doe@example.com"
        })

    def test_no_from_field(self):
        message_data = {}
        with self.assertRaises(AttributeError):
            getFromData(message_data)

    # Optionally: test for more edge cases, like "From": "<john.doe@example.com>" or other variation


class TestGetFormattedDate(unittest.TestCase):

    def test_standard_date(self):
        date = "Wed, 06 Jan 2021 12:34:56 +0000"
        result = getFormattedDate(date)
        self.assertEqual(result, "2021-01-06T12:34:56.000000Z")

    def test_date_with_utc_appendix(self):
        date = "Wed, 06 Jan 2021 12:34:56 +0000 (UTC)"
        result = getFormattedDate(date)
        self.assertEqual(result, "2021-01-06T12:34:56.000000Z")

    # def test_date_with_different_timezone(self):
    #     date = "Wed, 06 Jan 2021 14:34:56 +0200"
    #     result = getFormattedDate(date)
    #     # converting to UTC
    #     self.assertEqual(result, "2021-01-06T12:34:56.000000Z")

    def test_invalid_date_format(self):
        date = "2021-01-06 12:34:56"
        with self.assertRaises(ValueError):
            getFormattedDate(date)


class TestProcessMessageParts(unittest.TestCase):

    # TODO: Fix this test
    # @patch('client.valid_mime_type', return_value=True)
    # @patch('os.path.getsize', return_value=100)
    # def test_valid_parts(self, mock_getsize, mock_valid_mime_type):
    #     part1 = Mock()
    #     part1.get_content_maintype.return_value = 'text'
    #     part1.get.return_value = 'attachment'
    #     part1.get_filename.return_value = 'file1.txt'
    #     part1.get_content_type.return_value = 'text/plain'
    #     part1.get_payload.return_value = b'Hello, world!'

    #     part2 = Mock()
    #     part2.get_content_maintype.return_value = 'multipart'

    #     message_data = Mock()
    #     message_data.walk.return_value = [part1, part2]

    #     with tempfile.TemporaryDirectory() as temp_dir:
    #         # Update temp directory path
    #         global temp_dir_path
    #         temp_dir_path = temp_dir

    #         result = get_attachments(message_data)

    #     self.assertEqual(len(result), 1)
    #     self.assertEqual(result[0]['filename'], 'file1.txt')
    #     self.assertEqual(result[0]['fileType'], 'text/plain')
    #     self.assertEqual(result[0]['size'], 100)

    @patch('client.valid_mime_type', return_value=False)
    def test_invalid_mime_type(self, mock_valid_mime_type):
        part = Mock()
        part.get_content_maintype.return_value = 'text'
        part.get.return_value = 'attachment'
        part.get_filename.return_value = 'file1.txt'
        part.get_content_type.return_value = 'text/plain'

        message_data = Mock()
        message_data.walk.return_value = [part]

        result = get_attachments(message_data)

        self.assertEqual(len(result), 0)


if __name__ == '__main__':
    unittest.main()
