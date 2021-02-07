package main

import (
	"net"
	"testing"
	"time"

	"github.com/buildkite/interpolate"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/delamonpansie/silencer/config"
	"github.com/delamonpansie/silencer/filter"
)

func Test_banWorker(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ip1 := net.ParseIP("1.1.1.1").To4()
	ip2 := net.ParseIP("2.2.2.2").To4()
	ip3 := net.ParseIP("3.3.3.3").To4()

	blocker := filter.NewMockBlocker(ctrl)

	blocker.EXPECT().List().Return(nil)

	blocker.EXPECT().Block(ip1)
	blocker.EXPECT().Block(ip2)
	blocker.EXPECT().Block(ip3)

	var start time.Time
	blocker.EXPECT().Unblock(ip1).Do(func(_ interface{}) {
		delay := time.Now().Sub(start)
		assert.Less(t, time.Millisecond, delay)
		assert.Less(t, delay, 2*time.Millisecond)
	})
	blocker.EXPECT().Unblock(ip3).Do(func(_ interface{}) {
		delay := time.Now().Sub(start)
		assert.Less(t, time.Millisecond, delay)
		assert.Less(t, delay, 2*time.Millisecond)
	})
	blocker.EXPECT().Unblock(ip2).Do(func(_ interface{}) {
		delay := time.Now().Sub(start)
		assert.Less(t, 3*time.Millisecond, delay)
		assert.Less(t, delay, 4*time.Millisecond)
	})

	start = time.Now()
	c := worker(blocker, time.Second)
	c <- blockRequest{ip1, time.Millisecond}
	c <- blockRequest{ip1, time.Millisecond}
	c <- blockRequest{ip2, 3 * time.Millisecond}
	c <- blockRequest{ip3, time.Millisecond}

	time.Sleep(time.Millisecond * 10)
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

func Test_rule1(t *testing.T) {
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
	rule := testRule(t, "aaa", `($ip)`)
	m, err := rule.match("aaa bbb 1.1.1.1")
	require.NoError(t, err)
	assert.Equal(t, net.ParseIP("1.1.1.1").To4(), m)
}

func Test_rule22(t *testing.T) {
	rule := testRule(t, "aaa (.*)", " (.*)", `$ip`)
	m, err := rule.match("aaa bbb 1.1.1.1")
	require.NoError(t, err)
	assert.Equal(t, net.ParseIP("1.1.1.1").To4(), m)
}

func Test_rule3(t *testing.T) {
	rule := testRule(t, "zzz", `($ip)`)
	m, err := rule.match("aaa bbb 1.1.1.1")
	require.NoError(t, err)
	assert.Nil(t, m)
}
