import unittest

from utils import valid_from_email, valid_subject


class TestEmailUtils(unittest.TestCase):

    def test_valid_from_email_empty_whitelist(self):
        self.assertTrue(valid_from_email("test@example.com", []))

    def test_valid_from_email_in_whitelist(self):
        whitelist = [{"email": "test@example.com"}, {"email": "admin@example.com"}]
        self.assertTrue(valid_from_email("test@example.com", whitelist))

    def test_valid_from_email_not_in_whitelist(self):
        whitelist = [{"email": "admin@example.com"}]
        self.assertFalse(valid_from_email("test@example.com", whitelist))

    def test_valid_subject_empty_regex_list(self):
        self.assertTrue(valid_subject("Subject", []))

    def test_valid_subject_match_found(self):
        subject_line_regexes = [{"regex": r"^Test"}]
        self.assertTrue(valid_subject("Test subject", subject_line_regexes))

    def test_valid_subject_no_match_found(self):
        subject_line_regexes = [{"regex": r"^Hello"}]
        self.assertFalse(valid_subject("Test subject", subject_line_regexes))


if __name__ == "__main__":
    unittest.main()
