package wrk_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/nesvilab/philosopher/lib/sys"
	. "github.com/nesvilab/philosopher/lib/wrk"
)

var _ = Describe("Wrk", func() {

	Context("Testing workspace management", func() {

		It("Accessing workspace", func() {
			e := os.Chdir("../../test/wrksp")
			Expect(e).NotTo(HaveOccurred())
		})

		It("Init", func() {
			Init("0000", "0000")
		})

		It("Checking Meta Folder", func() {
			if _, e := os.Stat(sys.MetaDir()); os.IsNotExist(e) {
				Expect(e).NotTo(HaveOccurred())
			}
		})

		It("Checking Meta file", func() {
			if _, e := os.Stat(sys.Meta()); os.IsNotExist(e) {
				Expect(e).NotTo(HaveOccurred())
			}
		})

		It("Clean", func() {
			Clean()
		})

	})

})
