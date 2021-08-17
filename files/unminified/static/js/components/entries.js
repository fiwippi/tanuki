export const name = 'entries';

import * as Util from "/static/js/util.js"

export default function (entries, progress, urlFunc, extra) {
    return {
        search: "",
        images: Util.Images.BlankImageArray(entries.length),
        entries: Util.Ensure.Array(entries),
        progress: Util.Ensure.Object(progress),
        smallMedia: false,
        blankImage: Util.BlankImage,

        get filteredEntries() {
            return this.entries.filter(
                i => Util.Search.Match(this.search, i.title)
            )
        },

        imageWidth() {
            if (!this.smallMedia) return 200
            return (window.innerWidth * 0.8) / 2
        },

        getThumbnail(e) {
            if (e === undefined || this.images[e.order - 1] === undefined) {
                return this.blankImage
            }
            return this.images[e.order - 1].src
        },

        async init() {
            console.debug(this.entries)

            if (this.preInit !== undefined) {
                await this.preInit()
            }

            // Check if images need to be given smaller widths to fit into grid
            if (window.innerWidth <= Util.SmallMediaLimit) {
                this.smallMedia = true
            }

            await Util.Images.LoadImages(this.images, this.entries.length, urlFunc, true)

            if (this.postInit !== undefined) {
                await this.postInit()
            }
        },

        ...extra
    }
}