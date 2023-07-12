export const name = 'mangadex-search';

import * as API from "/static/js/api.js"
import * as Util from "/static/js/util.js"

export default function () {
    return {
        search: "",
        searchData: [],
        canShow: false,

        async searchEntries() {
            if (this.search.length === 0) {
                return []
            }

            this.canShow = false
            document.getElementById("spinner").classList.add("loader")

            await API.Mangadex.Search(this.search, 8)
                .then(data => {
                    this.searchData = data
                })

            document.getElementById("spinner").classList.remove("loader")
            this.canShow = true

            return this.searchData
        },

        fmtDescription(d) {
            if (d.length > 330) {
                return d.substring(0, 329) + "..."
            }
            return d
        },

        fmtDlLink(e) {
            return `/download/mangadex/${e.id}`
        },

        fmtCoverLink(e) {
            const parts = e.small_cover_url.split("/covers/")
            const endpoint = parts.at(parts.length - 1).replace("/", "_")
            return `/api/mangadex/cover/${endpoint}`
        },

        handleSearchChange(e) {
            this.searchData = []
        },
    }
}