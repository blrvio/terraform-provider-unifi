package acctest

import (
	"sync"
	"testing"

	pt "github.com/blrvio/terraform-provider-unifi/internal/provider/testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// settingUsgGeoLock serializes usg_geo setting tests (a site-wide singleton).
var settingUsgGeoLock = sync.Mutex{}

// TestAccSettingUsgGeo_basic covers the dedicated Region Blocking resource on
// Network 10.x, where Geo IP filtering lives in the separate `usg_geo` setting.
// It applies block/RU/both, asserts the values round-trip (enabled=true), and a
// second PlanOnly step requires an EMPTY plan (idempotency). Then it flips to
// allow to exercise the update path.
func TestAccSettingUsgGeo_basic(t *testing.T) {
	const res = "unifi_setting_usg_geo.test"

	AcceptanceTest(t, AcceptanceTestCase{
		VersionConstraint: ">= 9.0.0",
		Lock:              &settingUsgGeoLock,
		Steps: []resource.TestStep{
			{
				Config: testAccSettingUsgGeoConfigBlock(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(res, "ip_filtering.enabled", "true"),
					resource.TestCheckResourceAttr(res, "ip_filtering.action", "block"),
					resource.TestCheckResourceAttr(res, "ip_filtering.traffic_direction", "both"),
					resource.TestCheckResourceAttr(res, "ip_filtering.countries.#", "1"),
					resource.TestCheckResourceAttr(res, "ip_filtering.countries.0", "RU"),
				),
			},
			{
				Config:             testAccSettingUsgGeoConfigBlock(),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			pt.ImportStepWithSite(res),
			{
				Config: testAccSettingUsgGeoConfigAllow(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(res, "ip_filtering.enabled", "true"),
					resource.TestCheckResourceAttr(res, "ip_filtering.action", "allow"),
					resource.TestCheckResourceAttr(res, "ip_filtering.traffic_direction", "ingress"),
					resource.TestCheckResourceAttr(res, "ip_filtering.countries.#", "2"),
				),
			},
			pt.ImportStepWithSite(res),
		},
	})
}

func testAccSettingUsgGeoConfigBlock() string {
	return `
resource "unifi_site" "test" {
	description = "tfacc-setting-usg-geo"
}

resource "unifi_setting_usg_geo" "test" {
	site = unifi_site.test.name
	ip_filtering = {
		action    = "block"
		countries = ["RU"]
	}
}
`
}

func testAccSettingUsgGeoConfigAllow() string {
	return `
resource "unifi_site" "test" {
	description = "tfacc-setting-usg-geo"
}

resource "unifi_setting_usg_geo" "test" {
	site = unifi_site.test.name
	ip_filtering = {
		action            = "allow"
		countries         = ["US", "CA"]
		traffic_direction = "ingress"
	}
}
`
}
