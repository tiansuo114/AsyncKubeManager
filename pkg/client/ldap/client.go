package ldap

import (
	"asyncKubeManager/pkg/model"
	"crypto/tls"
	"fmt"
	"github.com/go-ldap/ldap/v3"
	"log"
	"time"
)

// LDAPClient holds the connection and configuration to interact with the LDAP server
type LDAPClient struct {
	conn *ldap.Conn
	opts *Options
}

// NewLDAPClient creates and returns a new LDAPClient, establishing the connection
func NewLDAPClient(opts *Options) (*LDAPClient, error) {
	// Attempt to establish a connection with the LDAP server
	ldapURL := fmt.Sprintf("ldaps://%s:%d", opts.Host, opts.Port)
	conn, err := ldap.DialURL(ldapURL, ldap.DialWithTLSConfig(&tls.Config{InsecureSkipVerify: true}))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to LDAP server: %v", err)
	}

	// Set the timeout for any LDAP operations
	conn.SetTimeout(30 * time.Second)

	// Try to bind using the provided credentials (assuming simple authentication)
	err = conn.Bind(opts.LDAPUserName, opts.LDAPPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to bind to LDAP server: %v", err)
	}

	// Return the new LDAP client
	return &LDAPClient{
		conn: conn,
		opts: opts,
	}, nil
}

// Close closes the LDAP connection
func (c *LDAPClient) Close() {
	if c.conn != nil {
		err := c.conn.Close()
		if err != nil {
			log.Printf("failed to close LDAP connection: %v", err)
		}
	}
}

// Search performs an LDAP search query
func (c *LDAPClient) Search(filter string, attributes []string) ([]*ldap.Entry, error) {
	searchRequest := ldap.NewSearchRequest(
		c.opts.BaseDN,
		ldap.ScopeWholeSubtree, // Search the entire subtree
		ldap.NeverDerefAliases, // Never dereference aliases
		0,                      // No limit on the number of entries
		0,                      // No time limit
		false,                  // Don't typesafe
		filter,                 // Filter to apply
		attributes,             // Attributes to fetch
		nil,                    // Controls (optional)
	)

	// Execute the search
	searchResult, err := c.conn.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("LDAP search failed: %v", err)
	}

	return searchResult.Entries, nil
}

// Modify performs an LDAP modify operation
func (c *LDAPClient) Modify(dn string, modifyRequest *ldap.ModifyRequest) error {
	// Perform the modify operation
	err := c.conn.Modify(modifyRequest)
	if err != nil {
		return fmt.Errorf("LDAP modify failed: %v", err)
	}
	return nil
}

// Add performs an LDAP add operation
func (c *LDAPClient) Add(entry *ldap.AddRequest) error {
	// Perform the add operation
	err := c.conn.Add(entry)
	if err != nil {
		return fmt.Errorf("LDAP add failed: %v", err)
	}
	return nil
}

// Delete performs an LDAP delete operation
func (c *LDAPClient) Delete(dn string) error {
	// Perform the delete operation
	delRequest := ldap.NewDelRequest(dn, nil)
	err := c.conn.Del(delRequest)
	if err != nil {
		return fmt.Errorf("LDAP delete failed: %v", err)
	}
	return nil
}

func (c *LDAPClient) FindUserByUID(uid string) (*model.LdapUser, error) {
	// 构造 LDAP 搜索请求
	searchRequest := ldap.NewSearchRequest(
		c.opts.BaseDN, // 基础 DN
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(uid=%s)", ldap.EscapeFilter(uid)),
		[]string{"dn", "cn", "ou", "uid", "sn", "givenName", "telephoneNumber", "mail", "gidNumber", "uidNumber", "homeDirectory", "userPassword"},
		nil,
	)

	// 执行搜索
	sr, err := c.conn.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("LDAP search failed: %w", err)
	}

	if len(sr.Entries) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	entry := sr.Entries[0]

	// 构造 LdapUser 实例
	user := &model.LdapUser{
		DN:              entry.DN,
		CN:              entry.GetAttributeValue("cn"),
		OU:              entry.GetAttributeValue("ou"),
		UID:             entry.GetAttributeValue("uid"),
		SN:              entry.GetAttributeValue("sn"),
		GivenName:       entry.GetAttributeValue("givenName"),
		TelephoneNumber: entry.GetAttributeValue("telephoneNumber"),
		Mail:            entry.GetAttributeValue("mail"),
		GIDNumber:       entry.GetAttributeValue("gidNumber"),
		UIDNumber:       entry.GetAttributeValue("uidNumber"),
		HomeDirectory:   entry.GetAttributeValue("homeDirectory"),
		UserPassword:    []byte(entry.GetRawAttributeValue("userPassword")),
	}

	return user, nil
}

func (c *LDAPClient) Bind(dn, password string) error {
	err := c.conn.Bind(dn, password)
	if err != nil {
		return fmt.Errorf("LDAP bind failed: %w", err)
	}
	return nil
}
