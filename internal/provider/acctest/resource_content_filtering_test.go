package acctest

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	pt "github.com/blrvio/terraform-provider-unifi/internal/provider/testing"
)

const testContentFilteringResourceName = "unifi_content_filtering.test"

// TestAccContentFiltering_basic is the regression lock for the create route that
// used to 405 (fixed by re-pointing go-unifi to POST content-filtering/create).
// It creates a per-network rule, requires a second apply to be an EMPTY plan
// (round-trip of categories/safe_search/network_ids), then exercises an update
// and import. Gated to 10.x where the content-filtering/create action exists.
func TestAccContentFiltering_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("tfacc-cf")
	subnet, vlanID := pt.GetTestVLAN(t)
	network := testAccContentFilteringNetwork(name, subnet.String(), vlanID)

	create := network + testAccContentFilteringConfig(name, `["ADVERTISEMENT"]`, `[]`)
	update := network + testAccContentFilteringConfig(name, `["ADVERTISEMENT", "FAMILY"]`, `["GOOGLE", "YOUTUBE", "BING"]`)

	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 10.0.0",
		Steps: []resource.TestStep{
			{
				Config: create,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testContentFilteringResourceName, "id"),
					resource.TestCheckResourceAttr(testContentFilteringResourceName, "name", name),
					resource.TestCheckResourceAttr(testContentFilteringResourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(testContentFilteringResourceName, "categories.#", "1"),
					resource.TestCheckResourceAttr(testContentFilteringResourceName, "network_ids.#", "1"),
				),
			},
			{
				// Round-trip: a second apply with no change must be an empty plan.
				Config:             create,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				Config: update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testContentFilteringResourceName, "categories.#", "2"),
					resource.TestCheckResourceAttr(testContentFilteringResourceName, "safe_search.#", "3"),
				),
			},
			pt.ImportStepWithSite(testContentFilteringResourceName),
		},
	})
}

func testAccContentFilteringNetwork(name, subnet string, vlanID int) string {
	return fmt.Sprintf(`
resource "unifi_network" "test" {
	name    = %[1]q
	purpose = "corporate"
	subnet  = %[2]q
	vlan_id = "%[3]d"
}
`, name, subnet, vlanID)
}

func testAccContentFilteringConfig(name, categories, safeSearch string) string {
	return fmt.Sprintf(`
resource "unifi_content_filtering" "test" {
	name        = %[1]q
	enabled     = true
	categories  = %[2]s
	safe_search = %[3]s
	network_ids = [unifi_network.test.id]

	schedule = {
		mode = "ALWAYS"
	}
}
`, name, categories, safeSearch)
}
