package pkg_test

import (
	"github.com/cirocosta/stacksearch/pkg"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Memory Dataset", func() {

	var (
		callstacksToAdd []pkg.Callstack
		dataset         pkg.Dataset
	)

	BeforeEach(func() {
		dataset = pkg.NewMemory()
	})

	JustBeforeEach(func() {
		for _, stack := range callstacksToAdd {
			dataset.Add(stack)
		}
	})

	Describe("Funcs", func() {
		var (
			fns []string
		)

		JustBeforeEach(func() {
			fns = dataset.Funcs()
		})

		Context("without any stacks", func() {
			It("finds nothing", func() {
				Expect(fns).To(BeEmpty())
			})
		})

		Context("having somethnig", func() {

			BeforeEach(func() {
				callstacksToAdd = append(callstacksToAdd,
					pkg.NewCallstack([]string{
						"fn1", "fn2",
					}, nil),
				)
			})

			It("gives back all functions", func() {
				Expect(fns).ToNot(BeEmpty())
				Expect(fns).To(ConsistOf("fn1", "fn2"))
			})
		})
	})

	Describe("Get", func() {

		var (
			fn     string
			stacks []pkg.Callstack
		)

		JustBeforeEach(func() {
			stacks, _ = dataset.Get(fn)
		})

		Context("without any stacks", func() {
			It("finds nothing", func() {
				Expect(stacks).To(BeEmpty())
			})
		})

		Context("having stacks added", func() {
			BeforeEach(func() {
				callstacksToAdd = append(callstacksToAdd,
					pkg.NewCallstack([]string{
						"fn1", "fn2",
					}, nil),
				)

				fn = "fn1"
			})

			It("finds stacks", func() {
				Expect(stacks).ToNot(BeEmpty())
			})
		})

	})

})
