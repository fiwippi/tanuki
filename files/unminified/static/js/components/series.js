export const name = 'series';

import * as API from "/static/js/api.js"
import * as Util from "/static/js/util.js"
import createEntries from "/static/js/components/entries.js"

export default function (series, entries, progress, seriesDataModal, entryViewModal, entryDataModal) {
    let urlFunc = (i) => {
        return `/api/series/${series.hash}/entries/${entries[i].hash}/cover?thumbnail=true`
    }

    let extra = {
        entry: {}, // Data about the entry the user is viewing
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
            // Load series data
            this.seriesMetadata.title = this.series.title
            this.seriesMetadata.author = this.series.author
            this.seriesMetadata.date_released = this.series.date_released
        },

        // Util functions
        width(index) {
            return Util.Images.Width(this.images, index)
        },

        // Progress related
        fmtEntryProgress(p) {
            return Util.Fmt.Percent(p.current / p.total)
        },

        fmtEntryProgressLabel(p) {
            return `Progress: ${this.fmtEntryProgress(p)}`
        },

        // Series Data Modal
        sModImg: {},
        sModShowMetadataResult: false,
        sModShowCoverResult: false,
        sModShowProgressResult: false,
        sModMetadataResult: "",
        sModCoverResult: "",
        sModProgressResult: "",

        sModPatchDataBtn: {
            async ['@click']() { await this.patchSeriesData() },
        },
        sModMetadataResultSpan: {
            ['x-show']() { return this.sModShowMetadataResult },
            ['x-text']() { return this.sModMetadataResult },
            ['@click.away']() { this.sModShowMetadataResult = false },
        },
        sModCoverResultSpan: {
            ['x-show']() { return this.sModShowCoverResult },
            ['x-text']() { return this.sModCoverResult },
            ['@click.away']() { this.sModShowCoverResult = false },
        },
        sModProgressResultSpan: {
            ['x-show']() { return this.sModShowProgressResult },
            ['x-text']() { return this.sModProgressResult },
            ['@click.away']() { this.sModShowProgressResult = false },
        },

        async showSeriesDataModal() {
            if (this.sModImg.raw === undefined) {
                await this.refreshSeriesThumbnail(true)
            }
            this.sdMod.show()
        },

        async refreshSeriesProgress() {
            await API.Catalog.SeriesProgress(this.series.hash)
                .then(progress => {
                    this.progress = progress
                })
                .catch(() => { this.progress = [] })
        },

        async refreshSeriesThumbnail(forceNew) {
            let sid = this.series.hash

            let img = new Image()
            let p =  Util.Images.WaitForLoad(img)
            if (forceNew) {
                img.src = `/api/series/${sid}/cover?thumbnail=true&time=${new Date().getTime()}`
            } else {
                img.src = `/api/series/${sid}/cover?thumbnail=true`
            }
            await p
            this.sModImg = img
        },

        async patchSeriesProgress(mode) {
            let amount = (mode === "read") ? "100%" : "0%"
            await API.Catalog.PatchProgress(this.series.hash, "", amount)
                .then(() => {
                    this.sModProgressResult = "Success!"
                    this.refreshSeriesProgress()
                })
                .catch(() => { this.sModProgressResult = "Failed!" })
            this.sModShowProgressResult = true
        },

        async patchSeriesData() {
            await API.Catalog.PatchSeries(this.series.hash, this.seriesMetadata.title,
                this.seriesMetadata.author, this.seriesMetadata.date_released)
                .then(() => {
                    this.sModMetadataResult = "Success!"
                    this.series.title = this.seriesMetadata.title // Refresh the title on screen
                })
                .catch(() => { this.sModMetadataResult = "Failed!" })

            this.sModShowMetadataResult = true
        },

        async patchSeriesCover(fileList) {
            if (fileList.length > 0) {
                await API.Catalog.PatchSeriesCover(this.series.hash, fileList[0], fileList[0].name)
                    .then(() => {
                        this.sModCoverResult = "Success!"
                        this.refreshSeriesThumbnail(true)
                    })
                    .catch(() => { this.sModCoverResult = "Failed!" })
                this.sModShowCoverResult = true
            }
        },

        async deleteSeriesCover() {
            await API.Catalog.DeleteSeriesCover(this.series.hash)
                .then(() => {
                    this.sModCoverResult = "Success!"
                    this.refreshSeriesThumbnail(true)
                })
                .catch(() => { this.sModCoverResult = "Failed!" })
            this.sModShowCoverResult = true
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
            let p = this.progress[entry.order - 1]
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
            await API.Catalog.PatchProgress(this.series.hash, this.entry.hash, amount)
                .then(() => { this.evModEntryProgress = "Success!" })
                .catch(() => { this.evModEntryProgress = "Failed!" })
            this.evModShowEntryProgress = true

            // Refresh it's new value
            await API.Catalog.EntryProgress(this.series.hash, this.entry.hash)
                .then(p => {
                    for (let i in this.entries) {
                        if (this.entries[i].hash === this.entry.hash) {
                            this.progress[this.entries[i].order - 1] = p
                        }
                    }
                })
        },

        // Entry Data Modal
        edModImg: {},
        edModEntryMetadataResult: "",
        edModShowEntryMetadataResult: "",
        edModEntryCoverResult: "",
        edModShowEntryCoverResult: "",

        edModMetadataResultSpan: {
            ['x-show']() { return this.edModShowEntryMetadataResult },
            ['x-text']() { return this.edModEntryMetadataResult },
            ['@click.away']() { this.edModShowEntryMetadataResult = false },
        },
        edModCoverResultSpan: {
            ['x-show']() { return this.edModShowEntryCoverResult },
            ['x-text']() { return this.edModEntryCoverResult },
            ['@click.away']() { this.edModShowEntryCoverResult = false },
        },

        async showEntryDataModal() {
            if (this.edModImg.raw === undefined) {
                await this.refreshEntryThumbnail(true)
            }
            this.evMod.hide()
            this.edMod.show()
        },

        async refreshEntryThumbnail(forceNew) {
            let sid = this.series.hash
            let eid = this.entry.hash

            let img = new Image()
            let p =  Util.Images.WaitForLoad(img)
            if (forceNew) {
                img.src = `/api/series/${sid}/entries/${eid}/cover?thumbnail=true&time=${new Date().getTime()}`
            } else {
                img.src = `/api/series/${sid}/entries/${eid}/cover?thumbnail=true`
            }
            await p
            this.edModImg = img

            for (let i in this.entries) {
                if (this.entries[i].hash === this.entry.hash) {
                    this.images[i] = img
                }
            }
        },

        async patchEntryData() {
            await API.Catalog.PatchEntry(
                this.series.hash,
                this.entry.hash,
                this.entryMetadata.title,
                this.entryMetadata.author,
                this.entryMetadata.date_released,
                this.entryMetadata.chapter,
                this.entryMetadata.volume
            )
                .then(() => {
                    this.edModEntryMetadataResult = "Success!"

                    // Refresh the local entry data
                    for (let i in this.entries) {
                        if (this.entries[i].hash === this.entry.hash) {
                            this.entries[i].title = this.entryMetadata.title
                        }
                    }
                })
                .catch(() => { this.edModEntryMetadataResult = "Failed!" })

            this.edModShowEntryMetadataResult = true
        },

        async patchEntryCover(fileList) {
            if (fileList.length > 0) {
                await API.Catalog.PatchEntryCover(this.series.hash, this.entry.hash, fileList[0], fileList[0].name)
                    .then(() => {
                        this.edModEntryCoverResult = "Success!"
                        this.refreshEntryThumbnail(true)
                    })
                    .catch(() => { this.edModEntryCoverResult = "Failed!" })

                this.edModShowEntryCoverResult = true
            }
        },

        async deleteEntryCover() {
            await API.Catalog.DeleteEntryCover(this.series.hash, this.entry.hash)
                .then(() => {
                    this.edModEntryCoverResult = "Success!"
                    this.refreshEntryThumbnail(true)
                })
                .catch(() => { this.edModEntryCoverResult = "Failed!" })

            this.edModShowEntryCoverResult = true
        },

        // Tag Related
        newTag: "",
        tagsPatchAllowed: false, // Stops x-effect with initial load of tags to send a patch

        sortedTags() {
            if (this.series.tags === undefined) {
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

            let sid = this.series.hash
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

    return createEntries(entries, progress, urlFunc, extra)
}