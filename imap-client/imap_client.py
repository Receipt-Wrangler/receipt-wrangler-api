from imapclient import IMAPClient


class ImapClient:
    host = None
    port = None
    username = None
    password = None
    client = None

    def __init__(self, host, port, username, password):
        self.host = host
        self.port = port
        self.username = username
        self.password = password
        self.client = IMAPClient(host=self.host, port=self.port)

    def connect(self):
        self.client.login(self.username, self.password)
