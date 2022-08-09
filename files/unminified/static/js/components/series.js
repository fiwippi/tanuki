export const name = 'series';

import * as API from "/static/js/api.js"
import * as Util from "/static/js/util.js"
import createEntries from "/static/js/components/entries.js"

export default function (series, entries, progress, seriesDataModal, entryViewModal, entryDataModal) {
    let urlFunc = (i) => {
        return `/api/series/${series.sid}/entries/${entries[i].eid}/cover?thumbnail=true`
    }
    let idFunc = (i) => {
        return entries[i].eid
    }

    let extra = {
        entry: {archive: {path: ""}}, // Data about the entry the user is viewing
        entryMetadata: {},
        series: Util.Ensure.Object(series), // Data about the series
        seriesMetadata: {}, // Holds metadata about the series,
                            // is separate from series data so
                            // when it's edited, it does not
                            // affect the content on the page

        // Modals
        sdMod: seriesDataModal,
        evMod: entryViewModal,
        edMod: entryDataModal,

        // Init related
        async preInit() {
            this.series.tags = Util.Ensure.Array(this.series.tags)
        },

        // Util functions
        width(index) {
            return Util.Images.Width(this.images, index)
        },

        // Progress related
        fmtEntryProgress(p) {
            return Util.Fmt.EntryPercent(p)
        },

        fmtEntryProgressLabel(p) {
            return `Progress: ${this.fmtEntryProgress(p)}`
        },

        // Series Data Modal
        sModShowProgressResult: false,
        sModProgressResult: "",

        sModProgressResultSpan: {
            ['x-show']() { return this.sModShowProgressResult },
            ['x-text']() { return this.sModProgressResult },
            ['@click.away']() { this.sModShowProgressResult = false },
        },

        async showSeriesDataModal() {
            this.sdMod.show()
        },

        async refreshSeriesProgress() {
            await API.Catalog.SeriesProgress(this.series.sid)
                .then(progress => {
                    this.progress = progress
                })
                .catch(() => { this.progress = [] })
        },

        async patchSeriesProgress(mode) {
            let amount = (mode === "read") ? "100%" : "0%"
            await API.Catalog.PatchProgress(this.series.sid, "", amount)
                .then(() => {
                    this.sModProgressResult = "Success!"
                    this.refreshSeriesProgress()
                })
                .catch(() => { this.sModProgressResult = "Failed!" })
            this.sModShowProgressResult = true
        },

        // Entry View Modal
        evModShowEntryProgress: false,
        evModEntryProgress: "",

        evModProgressResultSpan: {
            ['x-show']() { return this.evModShowEntryProgress },
            ['x-text']() { return this.evModEntryProgress },
            ['@click.away']() { this.evModShowEntryProgress = false },
        },

        entryPageProgress() {
            let entry = this.entry

            if (entry === undefined) {
                return 0
            }
            let p = this.progress[entry.eid]
            if (p === undefined) {
                return 0
            }
            return p.current
        },

        async showEntryViewModal(entry) {
            this.entry = entry
            this.entryMetadata.title = entry.title
            this.entryMetadata.author = entry.author
            this.entryMetadata.date_released = entry.date_released
            this.entryMetadata.chapter = entry.chapter
            this.entryMetadata.volume = entry.volume
            this.evMod.show()
        },

        async patchEntryProgress(mode) {
            // Set the progress
            let amount = (mode === "read") ? "100%" : "0%"
            await API.Catalog.PatchProgress(this.series.sid, this.entry.eid, amount)
                .then(() => { this.evModEntryProgress = "Success!" })
                .catch(() => { this.evModEntryProgress = "Failed!" })
            this.evModShowEntryProgress = true

            // Refresh it's new value
            await API.Catalog.EntryProgress(this.series.sid, this.entry.eid)
                .then(p => {
                    this.progress[this.entry.eid] = p
                })
        },

        // Tag Related
        newTag: "",
        tagsPatchAllowed: false, // Stops x-effect with initial load of tags to send a patch

        sortedTags() {
            if (this.series.tags === undefined || this.series.tags === null) {
                return []
            }
            return this.series.tags.sort(Util.Compare.Strings)
        },

        filterTags(tag) {
            this.tagsPatchAllowed = true
            this.series.tags = this.series.tags.filter(i => i !== tag)
        },

        async patchTags(tags) {
            if (!this.tagsPatchAllowed) {
                return
            }

            let sid = this.series.sid
            if (sid === undefined || sid.length === 0) {
                return
            }

            await API.Catalog.PatchTags(sid, tags)
        },

        addTag() {
            this.tagsPatchAllowed = true
            let newTag = this.newTag.trim()
            if (newTag !== "" && !this.series.tags.includes(newTag)) {
                this.series.tags.push(newTag);
                this.newTag = ""
            }
        },
    }

    return createEntries(entries, progress, urlFunc, idFunc, extra)
}