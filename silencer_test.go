package main

import (
	"net"
	"testing"
	"time"

	"github.com/buildkite/interpolate"
	"github.com/hpcloud/tail"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/delamonpansie/silencer/config"
	"github.com/delamonpansie/silencer/filter"
)

func testLogger(t *testing.T) {
	old := log
	log = zaptest.NewLogger(t)
	t.Cleanup(func() {
		log = old
	})
}
func testRule(t *testing.T, re ...string) rule {
	cfg := config.Load()
	env := interpolate.NewMapEnv(cfg.Env)
	// expand regex like "config.Load()" does
	for i := range re {
		s, err := interpolate.Interpolate(env, re[i])
		require.NoError(t, err)
		re[i] = s
	}
	return newRule("testRule", re, time.Second)
}

func Test_run_honors_whitelist(t *testing.T) {
	testLogger(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ip1 := net.ParseIP("192.168.1.1").To4()
	ip2, subnet, err := net.ParseCIDR("192.168.0.1/24")
	require.NoError(t, err)
	ip2 = ip2.To4()

	blocker := filter.NewMockBlocker(ctrl)
	blocker.EXPECT().Block(ip1, time.Second)

	whitelist := []net.IPNet{*subnet}
	lines := make(chan *tail.Line)
	go func() {
		lines <- &tail.Line{Text: ip2.String()}
		lines <- &tail.Line{Text: "192.168.0.120"}
		lines <- &tail.Line{Text: ip1.String()}
		close(lines)
	}()
	rules := []rule{testRule(t, "(.*)")}
	run(zap.NewNop(), blocker, lines, rules, whitelist)
}

func Test_rule1(t *testing.T) {
	testLogger(t)
	rule := testRule(t,
		`^$date_time \[\d+\] SMTP protocol error in "AUTH LOGIN" (.*)`,
		`(.*) AUTH command used when not advertised$`,
		`^H=\(\S+\) \[($ip)\]`,
	)

	line := `2020-10-27 21:05:50.780 [2168] SMTP protocol error in "AUTH LOGIN" H=(User) [103.253.42.54]:57715 I=[85.218.130.46]:25 AUTH command used when not advertised
`
	m, err := rule.match(line)
	require.NoError(t, err)
	assert.Equal(t, net.ParseIP("103.253.42.54").To4(), m)
}

func Test_rule2(t *testing.T) {
	testLogger(t)
	rule := testRule(t, "aaa", `($ip)`)
	m, err := rule.match("aaa bbb 1.1.1.1")
	require.NoError(t, err)
	assert.Equal(t, net.ParseIP("1.1.1.1").To4(), m)
}

func Test_rule22(t *testing.T) {
	testLogger(t)
	rule := testRule(t, "aaa (.*)", " (.*)", `$ip`)
	m, err := rule.match("aaa bbb 1.1.1.1")
	require.NoError(t, err)
	assert.Equal(t, net.ParseIP("1.1.1.1").To4(), m)
}

func Test_rule3(t *testing.T) {
	testLogger(t)
	rule := testRule(t, "zzz", `($ip)`)
	m, err := rule.match("aaa bbb 1.1.1.1")
	require.NoError(t, err)
	assert.Nil(t, m)
}
