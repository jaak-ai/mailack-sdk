import io
import unittest
from email.message import Message
from unittest.mock import patch
from urllib.error import HTTPError
from mailack import Client, APIError

class RawTests(unittest.TestCase):
    def test_downloads(self):
        for method, path, header in [
            ('get_message_raw', '/v1/messages/m/raw', 'X-Mailack-Canonical-Hash'),
            ('get_event_raw', '/v1/messages/m/events/e/raw', 'X-Mailack-Raw-SHA256'),
        ]:
            with self.subTest(method=method):
                response = io.BytesIO(b'\x00\xff\r\n\x80')
                response.status = 200
                response.headers = Message()
                response.headers[header] = 'digest'
                with patch('urllib.request.urlopen', return_value=response) as request:
                    args = ('m', 'e') if method == 'get_event_raw' else ('m',)
                    self.assertEqual(getattr(Client('https://example.test', 'test'), method)(*args),
                                     (b'\x00\xff\r\n\x80', 'digest'))
                    req = request.call_args.args[0]
                    self.assertEqual(req.full_url, 'https://example.test' + path)
                    self.assertEqual(req.get_header('Authorization'), 'Bearer test')
                    self.assertEqual(req.get_method(), 'GET')
    def test_errors(self):
        for method, args in [('get_message_raw', ('m',)), ('get_event_raw', ('m', 'e'))]:
            error = HTTPError('url', 404, 'missing', {}, io.BytesIO(b'{"error":{"code":"not_found"}}'))
            with patch('urllib.request.urlopen', side_effect=error):
                with self.assertRaises(APIError) as caught:
                    getattr(Client('https://example.test'), method)(*args)
                self.assertEqual(caught.exception.code, 'not_found')

    def test_json_regressions(self):
        for method, verb, suffix in [('seal_message', 'POST', 'seal'),
                                      ('message_evidence', 'GET', 'evidence'),
                                      ('proof_bundle', 'GET', 'proof-bundle')]:
            response = io.BytesIO(b'{"canonical_hash":"digest"}')
            response.status, response.headers = 200, Message()
            with patch('urllib.request.urlopen', return_value=response) as request:
                self.assertEqual(getattr(Client('https://example.test'), method)('m'),
                                 {'canonical_hash': 'digest'})
                req = request.call_args.args[0]
                self.assertEqual(req.full_url, 'https://example.test/v1/messages/m/' + suffix)
                self.assertEqual(req.get_method(), verb)
