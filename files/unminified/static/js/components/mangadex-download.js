export const name = 'mangadex-dl';

import * as API from "/static/js/api.js"
import * as MAPI from "/static/js/mangadex.js"
import * as Util from "/static/js/util.js"

export default function (uuid) {
    return {
        resp: "",
        search: "",
        uuid: uuid,
        data: {},
        chapters: [],
        checkboxes: [],
        createSub: false,

        get filteredChapters() {
            return this.chapters.filter(
                i => {
                    let a = Util.Search.Match(this.search, i.title)
                    let b = Util.Search.Match(this.search, i.chapter_no)
                    let c = Util.Search.Match(this.search, i.volume_no)
                    let d = Util.Search.Match(this.search, i.scanlation_group)

                    return (a || b || c || d)
                }
            )
        },

        selectAll(e) {
            this.checkboxes.fill(e.target.checked)
        },

        async downloadChapters() {
            this.resp = ""

            let chapters = []
            for (let i = 0; i < this.checkboxes.length; i++) {
                if (this.checkboxes[i] === true) {
                    chapters.push(this.chapters[i])
                }
            }

            await API.Download.Chapters(this.data.title, chapters, this.createSub)
                .then(() => {
                    this.resp = "Queued chapters"
                })
                .catch(() => {
                    this.resp = "Failed to queue chapters"
                })
        },

        fmtDate(d) {
            return Util.Fmt.RFCDate(d)
        },

        async init() {
            document.getElementById("spinner").classList.add("loader")

            await API.Mangadex.View(this.uuid)
                .then(resp => {
                    this.data = resp.listing
                    this.chapters = resp.chapters
                })

            this.checkboxes = new Array(this.chapters.length).fill(false)

            document.getElementById("spinner").classList.remove("loader")

            this.$watch('checkboxes', value => {
                let disabled = value.filter((i) => { return i === true }).length === 0
                document.getElementById('downloadButton').disabled = disabled
            })
        },
    }
}