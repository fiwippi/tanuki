export const name = 'entries';

import * as Util from "/static/js/util.js"

export default function (entries, progress, urlFunc, extra) {
    return {
        search: "",
        images: Util.Images.BlankImageArray(entries.length),
        entries: Util.Ensure.Array(entries),
        progress: Util.Ensure.Object(progress),
        smallMedia: false,

        get filteredEntries() {
            return this.entries.filter(
                i => Util.Search.Match(this.search, i.title)
            )
        },

        imageWidth() {
            if (!this.smallMedia) return 200
            return (window.innerWidth * 0.8) / 2
        },

        async init() {
            if (this.preInit !== undefined) {
                await this.preInit()
            }

            // Check if images need to be given smaller widths to fit into grid
            if (window.innerWidth <= Util.SmallMediaLimit) {
                this.smallMedia = true
            }

            let images = []
            await Util.Images.LoadImages(images, this.entries.length, urlFunc)
            this.images = images

            if (this.postInit !== undefined) {
                await this.postInit()
            }
        },

        ...extra
    }
}