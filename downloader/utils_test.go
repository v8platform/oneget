package downloader

import (
	"strings"
	"testing"
)

func TestMatchFilter_Match(t *testing.T) {

	type args struct {
		filters string
		project string
	}

	type want struct {
		source string
		want   bool
	}

	tests := []struct {
		filters string
		project string
		source  string
		want    bool
	}{
		{
			"server.x64.win",
			Platform83Project,
			"Cервер 1С:Предприятия (64-bit) для Windows",
			true,
		}, {
			"server.x64.win",
			Platform83Project,
			`Тонкий клиент 1С:Предприятие (64-bit) для Windows
Тонкий клиент 1С:Предприятия для Windows
Тонкий клиент 1С:Предприятие (64-bit) для Windows
Тонкий клиент 1С:Предприятия для DEB-based Linux-систем
Тонкий клиент 1С:Предприятия для RPM-based Linux-систем
Тонкий клиент 1С:Предприятия (64-bit) для DEB-based Linux-систем
Тонкий клиент 1С:Предприятия (64-bit) для RPM-based Linux-систем
Тонкий клиент 1С:Предприятия для macOS
Технологическая платформа 1С:Предприятия для Windows
Технологическая платформа 1С:Предприятия (64-bit) для Windows
Cервер 1С:Предприятия для DEB-based Linux-систем
Клиент 1С:Предприятия для macOS
Cервер 1С:Предприятия для RPM-based Linux-систем
Cервер 1С:Предприятия (64-bit) для DEB-based Linux-систем
Cервер 1С:Предприятия (64-bit) для RPM-based Linux-систем
Cервер 1С:Предприятия для Windows
`,
			false,
		}, {
			"server.x32.win",
			Platform83Project,
			"Cервер 1С:Предприятия для Windows",
			true,
		}, {
			"server.x32.win",
			Platform83Project,
			`Тонкий клиент 1С:Предприятие (64-bit) для Windows
Тонкий клиент 1С:Предприятия для Windows
Тонкий клиент 1С:Предприятие (64-bit) для Windows
Тонкий клиент 1С:Предприятия для DEB-based Linux-систем
Тонкий клиент 1С:Предприятия для RPM-based Linux-систем
Тонкий клиент 1С:Предприятия (64-bit) для DEB-based Linux-систем
Тонкий клиент 1С:Предприятия (64-bit) для RPM-based Linux-систем
Тонкий клиент 1С:Предприятия для macOS
Технологическая платформа 1С:Предприятия для Windows
Технологическая платформа 1С:Предприятия (64-bit) для Windows
Cервер 1С:Предприятия для DEB-based Linux-систем
Клиент 1С:Предприятия для macOS
Cервер 1С:Предприятия для RPM-based Linux-систем
Cервер 1С:Предприятия (64-bit) для DEB-based Linux-систем
Cервер 1С:Предприятия (64-bit) для RPM-based Linux-систем
`,
			false,
		}, {
			"server.win",
			Platform83Project,
			"Cервер 1С:Предприятия для Windows",
			true,
		}, {
			"server.win",
			Platform83Project,
			`Тонкий клиент 1С:Предприятие (64-bit) для Windows
Тонкий клиент 1С:Предприятия для Windows
Тонкий клиент 1С:Предприятие (64-bit) для Windows
Тонкий клиент 1С:Предприятия для DEB-based Linux-систем
Тонкий клиент 1С:Предприятия для RPM-based Linux-систем
Тонкий клиент 1С:Предприятия (64-bit) для DEB-based Linux-систем
Тонкий клиент 1С:Предприятия (64-bit) для RPM-based Linux-систем
Тонкий клиент 1С:Предприятия для macOS
Технологическая платформа 1С:Предприятия для Windows
Технологическая платформа 1С:Предприятия (64-bit) для Windows
Cервер 1С:Предприятия для DEB-based Linux-систем
Клиент 1С:Предприятия для macOS
Cервер 1С:Предприятия для RPM-based Linux-систем
Cервер 1С:Предприятия (64-bit) для DEB-based Linux-систем
Cервер 1С:Предприятия (64-bit) для RPM-based Linux-систем
`,
			false,
		}, {
			"thin.win",
			Platform83Project,
			"Тонкий клиент 1С:Предприятия для Windows",
			true,
		}, {
			"thin.win",
			Platform83Project,
			`Тонкий клиент 1С:Предприятие (64-bit) для Windows
Тонкий клиент 1С:Предприятие (64-bit) для Windows
Тонкий клиент 1С:Предприятия для DEB-based Linux-систем
Тонкий клиент 1С:Предприятия для RPM-based Linux-систем
Тонкий клиент 1С:Предприятия (64-bit) для DEB-based Linux-систем
Тонкий клиент 1С:Предприятия (64-bit) для RPM-based Linux-систем
Тонкий клиент 1С:Предприятия для macOS
Технологическая платформа 1С:Предприятия для Windows
Технологическая платформа 1С:Предприятия (64-bit) для Windows
Cервер 1С:Предприятия для DEB-based Linux-систем
Клиент 1С:Предприятия для macOS
Cервер 1С:Предприятия для RPM-based Linux-систем
Cервер 1С:Предприятия (64-bit) для DEB-based Linux-систем
Cервер 1С:Предприятия (64-bit) для RPM-based Linux-систем
Cервер 1С:Предприятия для Windows
`,
			false,
		}, {
			"win",
			Platform83Project,
			`Технологическая платформа 1С:Предприятия для Windows`,
			true,
		}, {
			"win.x64",
			Platform83Project,
			`Технологическая платформа 1С:Предприятия (64-bit) для Windows`,
			true,
		}, {
			"win",
			Platform83Project,
			`Тонкий клиент 1С:Предприятие (64-bit) для Windows
Тонкий клиент 1С:Предприятие (64-bit) для Windows
Тонкий клиент 1С:Предприятия для DEB-based Linux-систем
Тонкий клиент 1С:Предприятия для RPM-based Linux-систем
Тонкий клиент 1С:Предприятия (64-bit) для DEB-based Linux-систем
Тонкий клиент 1С:Предприятия (64-bit) для RPM-based Linux-систем
Тонкий клиент 1С:Предприятия для macOS
Тонкий клиент 1С:Предприятия для Windows
Cервер 1С:Предприятия для Windows
Технологическая платформа 1С:Предприятия (64-bit) для Windows
Cервер 1С:Предприятия для DEB-based Linux-систем
Клиент 1С:Предприятия для macOS
Cервер 1С:Предприятия для RPM-based Linux-систем
Cервер 1С:Предприятия (64-bit) для DEB-based Linux-систем
Cервер 1С:Предприятия (64-bit) для RPM-based Linux-систем
`,
			false,
		}, {
			"win.full",
			Platform83Project,
			`Технологическая платформа 1С:Предприятия для Windows`,
			true,
		}, {
			"win.full.x64",
			Platform83Project,
			`Технологическая платформа 1С:Предприятия (64-bit) для Windows`,
			true,
		}, {
			"win.full",
			Platform83Project,
			`Тонкий клиент 1С:Предприятие (64-bit) для Windows
Тонкий клиент 1С:Предприятие (64-bit) для Windows
Тонкий клиент 1С:Предприятия для DEB-based Linux-систем
Тонкий клиент 1С:Предприятия для RPM-based Linux-систем
Тонкий клиент 1С:Предприятия (64-bit) для DEB-based Linux-систем
Тонкий клиент 1С:Предприятия (64-bit) для RPM-based Linux-систем
Тонкий клиент 1С:Предприятия для macOS
Технологическая платформа 1С:Предприятия (64-bit) для Windows
Cервер 1С:Предприятия для DEB-based Linux-систем
Клиент 1С:Предприятия для macOS
Cервер 1С:Предприятия для RPM-based Linux-систем
Cервер 1С:Предприятия (64-bit) для DEB-based Linux-систем
Cервер 1С:Предприятия (64-bit) для RPM-based Linux-систем
`,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.filters, func(t *testing.T) {
			m, err := NewFilter(tt.project, tt.filters)

			if err != nil {
				t.Fatal(err)
			}

			cases := strings.Split(tt.source, "\n")
			for _, testCase := range cases {
				if got := m.MatchString(testCase); got != tt.want {
					t.Errorf("Match() = %v, want %v, source: %s", got, tt.want, testCase)
				}
			}

		})
	}
}
