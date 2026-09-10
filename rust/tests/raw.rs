use mailack::Client;
use std::io::{Read, Write};
use std::net::TcpListener;

#[tokio::test]
async fn raw_downloads() {
    for event in [false, true] {
        for status in [200, 404] {
            let listener = TcpListener::bind("127.0.0.1:0").unwrap();
            let url = format!("http://{}", listener.local_addr().unwrap());
            let server = std::thread::spawn(move || {
                let (mut stream, _) = listener.accept().unwrap();
                let mut request = Vec::new();
                loop {
                    let mut byte = [0];
                    stream.read_exact(&mut byte).unwrap();
                    request.push(byte[0]);
                    if request.ends_with(b"\r\n\r\n") { break; }
                }
                let request = String::from_utf8(request).unwrap();
                let path = if event { "/v1/messages/m/events/e/raw" } else { "/v1/messages/m/raw" };
                assert!(request.starts_with(&format!("GET {path} HTTP/1.1")));
                assert!(request.to_lowercase().contains("authorization: bearer test"));
                let header = if event { "x-mailack-raw-sha256" } else { "x-mailack-canonical-hash" };
                let body: &[u8] = if status == 200 { b"From: a@b\r\n\r\nraw" } else { br#"{"error":{"code":"not_found","message":"missing"}}"# };
                write!(stream, "HTTP/1.1 {status} OK\r\n{header}: digest\r\nContent-Length: {}\r\nConnection: close\r\n\r\n", body.len()).unwrap();
                stream.write_all(body).unwrap();
            });
            let client = Client::new(url).with_api_key("test");
            let result = if event {
                client.get_event_raw("m", "e").await.map(|r| (r.data, r.raw_sha256))
            } else {
                client.get_message_raw("m").await.map(|r| (r.data, r.canonical_hash))
            };
            server.join().unwrap();
            if status == 200 { assert_eq!(result.unwrap(), (b"From: a@b\r\n\r\nraw".to_vec(), "digest".into())); }
            else { assert!(result.unwrap_err().is_code("not_found")); }
        }
    }
}
