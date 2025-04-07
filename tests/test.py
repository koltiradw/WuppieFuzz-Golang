import socket
import requests
import unittest


HOST = "127.0.0.1"
PORT = 8080
LCOV_PORT = 3001
MTU = 4096

REQUEST_HEADER = bytearray([0x01, 0xC0, 0xC0, 0x10, 0x07])
COV_DUMP = bytearray([0x40, 0x01, 0x00])
COVERAGE_INFO_RESPONSE = 0x11

REF_REPORT = ['TN:', 'SF:test-rest-api/main.go', 'DA:28,1', 'DA:29,1', 'DA:30,1', 'DA:31,1', 'DA:32,1', 'DA:33,1', 
              'DA:34,1', 'DA:35,1', 'DA:38,1', 'DA:39,1', 'DA:40,1', 'DA:43,1', 'DA:44,1', 'DA:45,1', 'DA:46,1', 
              'DA:47,1', 'DA:48,1', 'DA:48,1', 'DA:49,1', 'DA:50,1', 'DA:53,1', 'DA:54,1', 'DA:59,1', 'DA:60,1', 
              'DA:61,1', 'DA:62,1', 'DA:63,1', 'DA:64,1', 'DA:64,1', 'DA:65,1', 'DA:65,1', 'DA:66,1', 'DA:67,1', 'DA:68,1', 'DA:70,1', 'end_of_record']

def send_request_to_agent(request: bytearray) -> bytearray:
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as stream:
        stream.settimeout(2)
        stream.connect((HOST, LCOV_PORT))
        stream.sendall(request)

        return stream.recv(MTU)

class TestGolangCovAgent(unittest.TestCase):
    def test_get_lcov(self):
        self.maxDiff = MTU
        requests.get(f"http://{HOST}:{PORT}/albums")
        requests.get(f"http://{HOST}:{PORT}/albums/1")
        requests.get(f"http://{HOST}:{PORT}/albums/4")
        requests.post(f"http://{HOST}:{PORT}/albums", json={"id": "4","title": "The Modern Sound of Betty Carter","artist": "Betty Carter","price": 49.99})
        requests.post(f"http://{HOST}:{PORT}/albums", json={"id": 5,"title": "The Modern Sound of Betty Carter","artist": "Betty Carter","price": 49.99})

        REQUEST_HEADER.extend(COV_DUMP) 
        agent_response = send_request_to_agent(REQUEST_HEADER)

        self.assertEqual(agent_response[0], COVERAGE_INFO_RESPONSE, "incorrect cov response flag")

        lcov_report = agent_response[5:].decode().split('\n')
        lcov_report.pop()
        
        self.assertEqual(REF_REPORT, lcov_report, "incorrect lcov report")

if __name__ == '__main__':
    unittest.main()