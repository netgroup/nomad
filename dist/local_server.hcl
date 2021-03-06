# Server has to run in a different endpoint, for example in your laptop

# Increase log verbosity
log_level = "DEBUG"

# Setup data dir
data_dir = "/tmp/local_server"

server {
	enabled = true

	# This is necessary for master election. In this case we have auto-proclamation
	bootstrap_expect = 1
}
