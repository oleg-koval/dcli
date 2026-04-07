package cmd

// MockDockerHelper implements DockerHelper for testing
type MockDockerHelper struct {
	GetServicesFn   func(projectDir string, profiles ...string) ([]string, error)
	RunCommandFn    func(projectDir string, args ...string) error
	GetContainersFn func() ([]string, error)
	Calls           struct {
		GetServices []struct {
			ProjectDir string
			Profiles   []string
		}
		RunCommand []struct {
			ProjectDir string
			Args       []string
		}
		GetContainers []struct{}
	}
}

func (m *MockDockerHelper) GetServices(projectDir string, profiles ...string) ([]string, error) {
	m.Calls.GetServices = append(m.Calls.GetServices, struct {
		ProjectDir string
		Profiles   []string
	}{projectDir, profiles})
	if m.GetServicesFn != nil {
		return m.GetServicesFn(projectDir, profiles...)
	}
	return []string{}, nil
}

func (m *MockDockerHelper) RunCommand(projectDir string, args ...string) error {
	m.Calls.RunCommand = append(m.Calls.RunCommand, struct {
		ProjectDir string
		Args       []string
	}{projectDir, args})
	if m.RunCommandFn != nil {
		return m.RunCommandFn(projectDir, args...)
	}
	return nil
}

func (m *MockDockerHelper) GetContainers() ([]string, error) {
	m.Calls.GetContainers = append(m.Calls.GetContainers, struct{}{})
	if m.GetContainersFn != nil {
		return m.GetContainersFn()
	}
	return []string{}, nil
}

// MockGitHelper implements GitHelper for testing
type MockGitHelper struct {
	IsGitRepoFn      func(path string) bool
	CheckoutBranchFn func(path, branch string) error
	ResetHardFn      func(path, branch string) error
	FetchOriginFn    func(path string) error
	Calls            struct {
		IsGitRepo      []struct{ Path string }
		CheckoutBranch []struct {
			Path   string
			Branch string
		}
		ResetHard []struct {
			Path   string
			Branch string
		}
		FetchOrigin []struct{ Path string }
	}
}

func (m *MockGitHelper) IsGitRepo(path string) bool {
	m.Calls.IsGitRepo = append(m.Calls.IsGitRepo, struct{ Path string }{path})
	if m.IsGitRepoFn != nil {
		return m.IsGitRepoFn(path)
	}
	return true
}

func (m *MockGitHelper) CheckoutBranch(path, branch string) error {
	m.Calls.CheckoutBranch = append(m.Calls.CheckoutBranch, struct {
		Path   string
		Branch string
	}{path, branch})
	if m.CheckoutBranchFn != nil {
		return m.CheckoutBranchFn(path, branch)
	}
	return nil
}

func (m *MockGitHelper) ResetHard(path, branch string) error {
	m.Calls.ResetHard = append(m.Calls.ResetHard, struct {
		Path   string
		Branch string
	}{path, branch})
	if m.ResetHardFn != nil {
		return m.ResetHardFn(path, branch)
	}
	return nil
}

func (m *MockGitHelper) FetchOrigin(path string) error {
	m.Calls.FetchOrigin = append(m.Calls.FetchOrigin, struct{ Path string }{path})
	if m.FetchOriginFn != nil {
		return m.FetchOriginFn(path)
	}
	return nil
}

// Helper functions for tests
func setDockerHelper(helper DockerHelper) {
	dockerHelper = helper
}

func resetDockerHelper() {
	dockerHelper = &defaultDockerHelper{}
}

func setGitHelper(helper GitHelper) {
	gitHelper = helper
}

func resetGitHelper() {
	gitHelper = &defaultGitHelper{}
}
