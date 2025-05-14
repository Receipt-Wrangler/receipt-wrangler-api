import unittest
from unittest.mock import patch

import client


class TestClient(unittest.TestCase):

    @patch('client.ImapClient')
    @patch('client.read_group_settings')
    @patch('client.init_logger')
    def test_main(self, mock_init_logger, mock_read_group_settings, mock_ImapClient):
        # Mocking the init_logger function
        mock_init_logger.return_value = None

        # Mocking the read_group_settings function
        mock_read_group_settings.return_value = [
            {
                "systemEmail": {
                    "id": "1",
                    "host": "imap.gmail.com",
                    "port": 993,
                    "username": "test@gmail.com",
                    "password": "password",
                    "useStartTLS": False
                },
                "subjectLineRegexes": [{"regex": ".*test.*"}],
                "emailWhiteList": [{"email": "test@example.com"}],
                "id": "1"
            }
        ]

        # Mocking the ImapClient class and its methods
        mock_client_instance = mock_ImapClient.return_value
        mock_client_instance.get_unread_email_metadata.return_value = [
            {
                "fromEmail": "test@example.com",
                "subject": "test subject",
                "groupSettingsIds": []
            }
        ]

        # Call the main function and catch the SystemExit exception
        try:
            client.main()
        except SystemExit as e:
            self.assertEqual(e.code, 0)

        # Assertions
        mock_init_logger.assert_called_once()
        mock_read_group_settings.assert_called_once()
        mock_ImapClient.assert_called_once_with(
            "imap.gmail.com",
            993,
            "test@gmail.com",
            "password",
            False,
            [{"regex": ".*test.*"}],
            [{"email": "test@example.com"}]
        )
        mock_client_instance.get_unread_email_metadata.assert_called_once()

    @patch('client.ImapClient')
    @patch('client.read_group_settings')
    @patch('client.init_logger')
    def test_main_no_system_emails(self, mock_init_logger, mock_read_group_settings, mock_ImapClient):
        # Mocking the init_logger function
        mock_init_logger.return_value = None

        # Mocking the read_group_settings function
        mock_read_group_settings.return_value = []

        # Call the main function and catch the SystemExit exception
        try:
            client.main()
        except SystemExit as e:
            self.assertEqual(e.code, 0)

        # Assertions
        mock_init_logger.assert_called_once()
        mock_read_group_settings.assert_called_once()
        mock_ImapClient.assert_not_called()

    @patch('client.ImapClient')
    @patch('client.read_group_settings')
    @patch('client.init_logger')
    def test_group_settings_ids(self, mock_init_logger, mock_read_group_settings, mock_ImapClient):
        # Mocking the init_logger function
        mock_init_logger.return_value = None

        # Mocking the read_group_settings function
        mock_read_group_settings.return_value = [
            {
                "systemEmail": {
                    "id": "1",
                    "host": "imap.gmail.com",
                    "port": 993,
                    "username": "test@gmail.com",
                    "password": "password",
                    "useStartTLS": False
                },
                "subjectLineRegexes": [{"regex": ".*test.*"}],
                "emailWhiteList": [{"email": "test@example.com"}],
                "id": "1"
            }
        ]

        # Mocking the ImapClient class and its methods
        mock_client_instance = mock_ImapClient.return_value
        mock_client_instance.get_unread_email_metadata.return_value = [
            {
                "fromEmail": "test@example.com",
                "subject": "test subject",
                "groupSettingsIds": []
            }
        ]

        # Call the main function and catch the SystemExit exception
        try:
            client.main()
        except SystemExit as e:
            self.assertEqual(e.code, 0)

        # Assertions
        mock_init_logger.assert_called_once()
        mock_read_group_settings.assert_called_once()
        mock_ImapClient.assert_called_once_with(
            "imap.gmail.com",
            993,
            "test@gmail.com",
            "password",
            False,
            [{"regex": ".*test.*"}],
            [{"email": "test@example.com"}]
        )
        mock_client_instance.get_unread_email_metadata.assert_called_once()
        self.assertEqual(mock_client_instance.get_unread_email_metadata.return_value[0]['groupSettingsIds'], ["1"])


if __name__ == '__main__':
    unittest.main()
