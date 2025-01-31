package messageWithCounter

//go:generate protoc --proto_path=. --go_out=. --go_opt=default_api_level=API_OPAQUE --go_opt=paths=source_relative messageWithCounter.proto
