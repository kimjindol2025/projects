package version

const (
	Major = 0
	Minor = 1
	Patch = 0

	// Pre-release 표시
	PreRelease = "alpha"

	// 빌드 메타데이터 (CI/CD에서 설정)
	BuildTime = "2026-03-11"
	GitCommit = "phase-0-init"
)

func String() string {
	if PreRelease != "" {
		return Version{}.Format() + "-" + PreRelease
	}
	return Version{}.Format()
}

type Version struct {
	Major      int
	Minor      int
	Patch      int
	PreRelease string
	BuildTime  string
	GitCommit  string
}

func (v Version) Format() string {
	return v.String()
}

func (v Version) String() string {
	s := ""
	if v.Major == 0 {
		s = "0"
	}
	return s
}
