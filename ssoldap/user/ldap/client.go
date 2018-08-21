package ldap

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mijia/sweb/log"
	"gopkg.in/ldap.v2"
)

var (
	ErrNoSuchObject = errors.New("no such object in ldap")
	ErrAuthFail     = errors.New("user or pwd wrong in ldap")
)

// LdapClient
type LdapClient struct {
	ldapHost        string
	ldapPort        int
	username        string
	password        string
	baseDn          string
	conn            *ldap.Conn
	lastConnectTime time.Time //上次重连时间点，这里不区分是否重连成功
	lock            *sync.Mutex
}

// NewLdapClient 返回一个暴露了查询方法的ldap客户端
// ldapHost 1.1.1.1
// port     389
// baseDn   OU=xxx,DC=xxx
func NewClient(ldapHost string, ldapPort int, baseDn string, username string, password string) (*LdapClient, error) {
	client := &LdapClient{
		ldapHost: ldapHost,
		ldapPort: ldapPort,
		username: username,
		password: password,
		baseDn:   baseDn,
		lock:     &sync.Mutex{},
	}

	// 下面为初始化代码，和重连代码分开
	err := client.refreshConn()
	if err != nil {
		return nil, err
	}

	return client, nil
}

// refreshConn 更新当前的连接, 该方法应该只在reconnect中和初始化时候被调用
func (c *LdapClient) refreshConn() error {
	conn, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", c.ldapHost, c.ldapPort))
	if err != nil {
		// todo 这里应该发送邮件报警
		log.Errorf("failed to connect ldap server, err %+v", err)
		return err
	}

	// First bind with a read only user
	err = conn.Bind(c.username, c.password)
	if err != nil {
		log.Errorf("failed to login ldap, err %+v", err)
		return err
	}
	c.conn = conn
	c.lastConnectTime = time.Now()
	log.Info("successfully refresh ldap connection")
	return nil
}

// reconnect 我们通过限制重连间隔来防止雪崩, 我们只有在实际发生了重连并且失败的情况下才会返回error
func (c *LdapClient) reconnect() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	// 我们只有在连接为空，或者上次重连超过5秒后才会进行重连
	if !(c.conn == nil || c.lastConnectTime.IsZero() || time.Now().After(c.lastConnectTime.Add(5*time.Second))) {
		log.Warn("ldap connection reconnect too often, will skip")
		return nil
	}
	return c.refreshConn()
}

// Close 关闭连接
func (c *LdapClient) Close() error {
	if c.conn != nil {
		c.conn.Close()
	}
	log.Info("successfully close ldap client")
	c.conn = nil
	c.lastConnectTime = time.Time{}
	return nil
}

func (c *LdapClient) search(searchRequest *ldap.SearchRequest) (*ldap.SearchResult, error) {
	sr, err := c.conn.Search(searchRequest)
	if err != nil {
		log.Errorf("failed to search ldap for request %+v, err %+v", searchRequest, err)
		// 这里需要判断是否是连接错误，如果是的话，那么就需要重连，然后重试
		if !strings.Contains(err.Error(), "Network Error") {
			return nil, err
		}
		if c.reconnect() == nil {
			sr, err = c.conn.Search(searchRequest)
			if err != nil {
				log.Errorf("failed to search ldap for request %+v again, will not retry, err %+v", searchRequest, err)
				return nil, err
			}
			return sr, nil
		}
		log.Errorf("failed to reconnect ldap for request %+v, will not retry, err %+v", searchRequest, err)
		return nil, err
	}
	return sr, nil
}

// SearchForUser 通过upn也就是邮箱来查询用户, todo 根据上层调用方使用情况，返回值可以做一定的包装
func (c *LdapClient) SearchForUser(upn string) (*ldap.SearchResult, error) {
	searchRequest := ldap.NewSearchRequest(
		c.baseDn,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		fmt.Sprintf("(userPrincipalName=%s)", upn),
		[]string{"cn", "userPrincipalName"},
		nil,
	)

	return c.search(searchRequest)
}

// SearchForOU  OUs名字类似于 OU=***,OU=***,OU=**,OU=**,DC=*,DC=*
// 这个方法其实很不实用, todo 根据上层调用方使用情况，返回值可以做一定的包装
func (c *LdapClient) SearchForOU(OUs string) (*ldap.SearchResult, error) {
	searchRequest := ldap.NewSearchRequest(
		OUs,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		"(&(objectClass=organizationalUnit)(objectClass=top))",
		[]string{"name", "ou"},
		nil,
	)

	sr, err := c.search(searchRequest)
	if err != nil {
		log.Errorf("failed to search ldap for request %+v, err %+v", searchRequest, err)
		if strings.Contains(err.Error(), "No Such Object") {
			return nil, ErrNoSuchObject
		}
		return nil, err
	}
	return sr, nil
}

// Auth 验证用户名密码， todo ut 测试， mail到底有没有后缀， 还有就是明文没有问题么
func (c *LdapClient) Auth(mail string, passwd string) (bool, error) {
	conn, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", c.ldapHost, c.ldapPort))
	if err != nil {
		// todo 这里应该发送邮件报警
		log.Errorf("failed to connect ldap server, err %+v", err)
		return false, err
	}

	defer conn.Close()

	// First bind with a read only user
	err = conn.Bind(mail, passwd)
	if err != nil {
		log.Errorf("failed to login ldap, err %+v", err)
		if strings.Contains(err.Error(), "Invalid Credentials") {
			return false, ErrAuthFail
		}
		return false, err
	}
	return true, nil

}
