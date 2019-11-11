package pkg_test

import (
	"regexp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/cirocosta/stacksearch/pkg"
	pprof "github.com/google/pprof/profile"
)

var _ = Describe("Stack", func() {

	Describe("NewCallstack", func() {
		var (
			callstack pkg.Callstack
			data      []string
			opt       *pkg.CallstackOptions
		)

		JustBeforeEach(func() {
			callstack = pkg.NewCallstack(data, nil, opt)
		})

		Context("with StopAt", func() {
			BeforeEach(func() {
				opt = &pkg.CallstackOptions{
					StopAt: regexp.MustCompile("^test.*"),
				}
			})

			Context("starting w/ fns not matching", func() {
				BeforeEach(func() {
					data = []string{
						"foo", "bar", "test.fn1", "test.fn2",
					}
				})

				It("cuts them out", func() {
					Expect(callstack.Data).To(ConsistOf(
						"test.fn1", "test.fn2",
					))
				})
			})
		})
	})

	Describe("MergeSubCallstacks", func() {

		var callstacks, merged []pkg.Callstack

		JustBeforeEach(func() {
			merged = pkg.MergeSubCallstacks(callstacks)
		})

		Context("with empty callstacks", func() {
			BeforeEach(func() {
				callstacks = nil
			})

			It("does nothing", func() {
				Expect(merged).To(Equal(callstacks))
			})
		})

		Context("with single callstack", func() {
			BeforeEach(func() {
				callstacks = []pkg.Callstack{}
			})

			It("does nothinig", func() {
				Expect(merged).To(BeEmpty())
			})
		})

		Context("with 2+ callstacks", func() {
			Context("being identical", func() {
				BeforeEach(func() {
					callstacks = []pkg.Callstack{
						{Data: []string{"fn1", "fn2"}},
						{Data: []string{"fn1", "fn2"}},
					}
				})

				It("reduces to a single one", func() {
					Expect(merged).To(ConsistOf(pkg.Callstack{
						Data: []string{"fn1", "fn2"},
					}))
				})
			})

			Context("havinig one that is part of another", func() {
				BeforeEach(func() {
					callstacks = []pkg.Callstack{
						{Data: []string{"fn1", "fn2"}},
						{Data: []string{"fn1"}},
					}
				})

				It("merges", func() {
					Expect(merged).To(ConsistOf(pkg.Callstack{
						Data: []string{"fn1", "fn2"},
					}))
				})
			})
		})
	})

	Describe("CallstacksFromPprof", func() {

		var (
			err        error
			callstacks []pkg.Callstack
			src        *pprof.Profile

			opt pkg.CallstackOptions
		)

		BeforeEach(func() {
			opt = pkg.CallstackOptions{}
		})

		JustBeforeEach(func() {
			callstacks, err = pkg.CallstacksFromPprof(src, opt)
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
			})

			Context("without verbose set", func() {
				It("doesn't capture file info", func() {
					Expect(callstacks[0].Locations).To(BeEmpty())
				})
			})

			Context("having verbose set", func() {
				BeforeEach(func() {
					opt = pkg.CallstackOptions{
						Verbose: true,
					}
				})

				It("has file info captured", func() {
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
					pkg.NewCallstack([]string{"a"}, nil, nil),
				},
				expected: []pkg.Callstack{
					pkg.NewCallstack([]string{"a"}, nil, nil),
				},
			}),
			Entry("different stacks", scenario{
				input: []pkg.Callstack{
					pkg.NewCallstack([]string{"a"}, nil, nil),
					pkg.NewCallstack([]string{"b"}, nil, nil),
				},
				expected: []pkg.Callstack{
					pkg.NewCallstack([]string{"a"}, nil, nil),
					pkg.NewCallstack([]string{"b"}, nil, nil),
				},
			}),
			Entry("equal stacks", scenario{
				input: []pkg.Callstack{
					pkg.NewCallstack([]string{"a"}, nil, nil),
					pkg.NewCallstack([]string{"a"}, nil, nil),
				},
				expected: []pkg.Callstack{
					pkg.NewCallstack([]string{"a"}, nil, nil),
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
