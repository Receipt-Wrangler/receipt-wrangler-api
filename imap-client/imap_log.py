import logging
import os
import sys

def create_log_file(path):
	if not os.path.exists(path):
		with open(path, "w") as f:
			pass

def init_logger():
    log_format = '%(asctime)s %(levelname)s {%(pathname)s:%(lineno)d} %(message)s'

    base_path = os.environ.get("BASE_PATH", "")
    path = os.path.join(base_path, "logs", "imap-client.log")

    create_log_file(path)

    stdout_handler = logging.StreamHandler(sys.stdout)
    file_handler = logging.FileHandler(filename=path)

    logging.basicConfig(
        format=log_format,
        handlers=[stdout_handler, file_handler]
    )
