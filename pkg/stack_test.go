package pkg_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/cirocosta/stacksearch/pkg"
	pprof "github.com/google/pprof/profile"
)

var _ = Describe("Stack", func() {

	Describe("CallstacksFromPprof", func() {

		var (
			err        error
			callstacks []pkg.Callstack
			src        *pprof.Profile
		)

		JustBeforeEach(func() {
			callstacks, err = pkg.CallstacksFromPprof(src)
		})

		Context("with empty profile", func() {
			BeforeEach(func() {
				src = nil
			})

			It("errors", func() {
				Expect(err).To(HaveOccurred())
			})
		})

		Context("having no samples", func() {
			BeforeEach(func() {
				src = &pprof.Profile{}
			})

			It("succeeds", func() {
				Expect(err).ToNot(HaveOccurred())
			})

			It("produces an empty set of callstacks", func() {
				Expect(callstacks).To(BeEmpty())
			})
		})

		Context("w/ samples", func() {
			BeforeEach(func() {
				src = &pprof.Profile{
					Sample: []*pprof.Sample{
						{
							Location: []*pprof.Location{
								location("fn1"),
								location("fn2"),
							},
						},
					},
				}
			})

			It("produces a callstack", func() {
				Expect(callstacks).To(HaveLen(1))
			})

			It("has function info captured", func() {
				Expect(callstacks[0].Data).To(ConsistOf(
					"fn1",
					"fn2",
				))

				Expect(callstacks[0].Locations).To(ConsistOf(
					pkg.Location{
						Filename: "fn1.go",
						Line:     123,
					},
					pkg.Location{
						Filename: "fn2.go",
						Line:     123,
					},
				))
			})
		})

	})

	Describe("Merge", func() {

		type scenario struct{ input, expected []pkg.Callstack }

		DescribeTable("having",
			func(s scenario) {
				Expect(pkg.Merge(s.input)).To(Equal(s.expected))
			},
			Entry("empty", scenario{}),
			Entry("single", scenario{
				input: []pkg.Callstack{
					pkg.NewCallstack([]string{"a"}, nil),
				},
				expected: []pkg.Callstack{
					pkg.NewCallstack([]string{"a"}, nil),
				},
			}),
			Entry("different stacks", scenario{
				input: []pkg.Callstack{
					pkg.NewCallstack([]string{"a"}, nil),
					pkg.NewCallstack([]string{"b"}, nil),
				},
				expected: []pkg.Callstack{
					pkg.NewCallstack([]string{"a"}, nil),
					pkg.NewCallstack([]string{"b"}, nil),
				},
			}),
			Entry("equal stacks", scenario{
				input: []pkg.Callstack{
					pkg.NewCallstack([]string{"a"}, nil),
					pkg.NewCallstack([]string{"a"}, nil),
				},
				expected: []pkg.Callstack{
					pkg.NewCallstack([]string{"a"}, nil),
				},
			}),
		)
	})

})

func location(fn string) *pprof.Location {
	return &pprof.Location{
		Line: []pprof.Line{
			{
				Function: &pprof.Function{
					Name:     fn,
					Filename: fn + ".go",
				},
				Line: 123,
			},
		},
	}
}
