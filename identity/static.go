package identity

func StaticService(port int) Identity {
	host := localPrivateHost()
	return &identity{
		private: &address{
			host: host,
			port: port,
		},
		public: &address{
			host: host,
			port: port,
		},
	}
}
