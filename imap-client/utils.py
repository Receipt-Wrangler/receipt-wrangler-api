import logging
import re


def valid_from_email(from_email, email_whitelist):
    return from_email in email_whitelist


def valid_subject(subject, subject_line_regexes):
    for subject_line_regex in subject_line_regexes:
        regex = re.compile(subject_line_regex["regex"])
        matches = regex.search(subject)
        logging.info(
            f"Found match: {matches} on email subject: '{subject}' with regex: '{subject_line_regex['regex']}'")
        if matches:
            return True

    return False
