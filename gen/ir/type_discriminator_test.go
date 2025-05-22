package ir

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSumSpec_GetAllDiscriminatorValues(t *testing.T) {
	// Create Type instances once to ensure pointer equality
	testType := &Type{Name: "Test"}
	userType := &Type{Name: "User"}
	adminType := &Type{Name: "Admin"}
	notFoundType := &Type{Name: "NotFound"}

	tests := []struct {
		name     string
		spec     SumSpec
		typ      *Type
		expected []string
	}{
		{
			name: "no discriminator",
			spec: SumSpec{
				Discriminator: "",
			},
			typ:      testType,
			expected: nil,
		},
		{
			name: "no mappings",
			spec: SumSpec{
				Discriminator: "type",
				Mapping:       []SumSpecMap{},
			},
			typ:      testType,
			expected: nil, // empty slice becomes nil when no results found
		},
		{
			name: "single mapping",
			spec: SumSpec{
				Discriminator: "type",
				Mapping: []SumSpecMap{
					{Key: "user", Type: userType},
				},
			},
			typ:      userType,
			expected: []string{"user"},
		},
		{
			name: "multiple mappings same type",
			spec: SumSpec{
				Discriminator: "type",
				Mapping: []SumSpecMap{
					{Key: "user.created", Type: userType},
					{Key: "user.updated", Type: userType},
					{Key: "admin.login", Type: adminType},
				},
			},
			typ:      userType,
			expected: []string{"user.created", "user.updated"},
		},
		{
			name: "multiple mappings different types",
			spec: SumSpec{
				Discriminator: "type",
				Mapping: []SumSpecMap{
					{Key: "user.created", Type: userType},
					{Key: "admin.login", Type: adminType},
					{Key: "user.updated", Type: userType},
				},
			},
			typ:      adminType,
			expected: []string{"admin.login"},
		},
		{
			name: "type not found",
			spec: SumSpec{
				Discriminator: "type",
				Mapping: []SumSpecMap{
					{Key: "user", Type: userType},
				},
			},
			typ:      notFoundType,
			expected: nil, // empty slice becomes nil when no results found
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.spec.GetAllDiscriminatorValues(tt.typ)
			if tt.expected == nil {
				require.Nil(t, result)
			} else {
				require.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestSumSpec_GetMapValue(t *testing.T) {
	userType := &Type{Name: "User"}
	adminType := &Type{Name: "Admin"}

	spec := SumSpec{
		Discriminator: "type",
		Mapping: []SumSpecMap{
			{Key: "user.created", Type: userType, FormattedName: "UserCreated"},
			{Key: "user.updated", Type: userType, FormattedName: "UserUpdated"},
			{Key: "admin.login", Type: adminType, FormattedName: "AdminLogin"},
		},
	}

	tests := []struct {
		name     string
		typ      *Type
		key      string
		expected SumSpecMap
		panics   bool
	}{
		{
			name: "valid mapping found",
			typ:  userType,
			key:  "user.created",
			expected: SumSpecMap{
				Key:           "user.created",
				Type:          userType,
				FormattedName: "UserCreated",
			},
		},
		{
			name: "another valid mapping",
			typ:  adminType,
			key:  "admin.login",
			expected: SumSpecMap{
				Key:           "admin.login",
				Type:          adminType,
				FormattedName: "AdminLogin",
			},
		},
		{
			name:   "mapping not found - wrong type",
			typ:    userType,
			key:    "admin.login",
			panics: true,
		},
		{
			name:   "mapping not found - wrong key",
			typ:    userType,
			key:    "nonexistent",
			panics: true,
		},
		{
			name:   "mapping not found - both wrong",
			typ:    &Type{Name: "Unknown"},
			key:    "nonexistent",
			panics: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.panics {
				require.Panics(t, func() {
					spec.GetMapValue(tt.typ, tt.key)
				})
			} else {
				result := spec.GetMapValue(tt.typ, tt.key)
				require.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestSumSpecMap_ValueGo(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "simple string",
			key:      "user",
			expected: `"user"`,
		},
		{
			name:     "string with dots",
			key:      "user.created",
			expected: `"user.created"`,
		},
		{
			name:     "empty string",
			key:      "",
			expected: `""`,
		},
		{
			name:     "string with quotes",
			key:      `test"quote`,
			expected: `"test\"quote"`,
		},
		{
			name:     "string with backslash",
			key:      `test\slash`,
			expected: `"test\\slash"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sumSpecMap := SumSpecMap{Key: tt.key}
			result := sumSpecMap.ValueGo()
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestSumSpecIntegration(t *testing.T) {
	// Test integration of all discriminator mapping functions together
	userType := &Type{Name: "UserEvent"}
	adminType := &Type{Name: "AdminEvent"}

	spec := SumSpec{
		Discriminator: "eventType",
		Mapping: []SumSpecMap{
			{Key: "user.created", Type: userType, FormattedName: "UserCreated"},
			{Key: "user.updated", Type: userType, FormattedName: "UserUpdated"},
			{Key: "user.deleted", Type: userType, FormattedName: "UserDeleted"},
			{Key: "admin.login", Type: adminType, FormattedName: "AdminLogin"},
			{Key: "admin.logout", Type: adminType, FormattedName: "AdminLogout"},
		},
	}

	t.Run("user event has multiple discriminator values", func(t *testing.T) {
		values := spec.GetAllDiscriminatorValues(userType)
		require.Len(t, values, 3)
		require.Contains(t, values, "user.created")
		require.Contains(t, values, "user.updated")
		require.Contains(t, values, "user.deleted")
	})

	t.Run("admin event has multiple discriminator values", func(t *testing.T) {
		values := spec.GetAllDiscriminatorValues(adminType)
		require.Len(t, values, 2)
		require.Contains(t, values, "admin.login")
		require.Contains(t, values, "admin.logout")
	})

	t.Run("can get specific mapping values", func(t *testing.T) {
		mapping := spec.GetMapValue(userType, "user.created")
		require.Equal(t, "user.created", mapping.Key)
		require.Equal(t, userType, mapping.Type)
		require.Equal(t, "UserCreated", mapping.FormattedName)
		require.Equal(t, `"user.created"`, mapping.ValueGo())
	})

	t.Run("unknown type returns empty values", func(t *testing.T) {
		unknownType := &Type{Name: "Unknown"}
		values := spec.GetAllDiscriminatorValues(unknownType)
		require.Empty(t, values)
	})
}
