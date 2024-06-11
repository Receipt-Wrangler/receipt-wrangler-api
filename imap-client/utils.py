import logging
import re


def valid_from_email(from_email, email_whitelist):
    if len(email_whitelist) == 0:
        return True

    whitelist_email_addresses = list(
        map(lambda emails: emails["email"], email_whitelist))
    if from_email not in whitelist_email_addresses:
        return False

    return True


def valid_subject(subject, subject_line_regexes):
    if len(subject_line_regexes) == 0:
        return True

    for subject_line_regex in subject_line_regexes:
        regex = re.compile(subject_line_regex["regex"])
        matches = regex.search(subject)
        logging.info(
            f"Found match: {matches} on email subject: '{subject}' with regex: '{subject_line_regex['regex']}'")
        if matches:
            return True

    return False
