export const name = 'entries';

import * as Util from "/static/js/util.js"

export default function (entries, progress, urlFunc, idFunc, extra) {
    return {
        search: "",
        images: new Map(),
        entries: Util.Ensure.Array(entries),
        progress: Util.Ensure.Object(progress),
        smallMedia: false,
        blankImage: Util.BlankImage,

        get filteredEntries() {
            return this.entries.filter(
                i => Util.Search.Match(this.search, i.display_title || i.folder_title || i.title)
            )
        },

        imageWidth() {
            if (!this.smallMedia) return 200
            return (window.innerWidth * 0.8) / 2
        },

        getThumbnail(id) {
            let img = this.images.get(id)
            if (img === undefined) {
                return this.blankImage
            }
            return img.src
        },

        async init() {
            if (this.preInit !== undefined) {
                await this.preInit()
            }

            // Check if images need to be given smaller widths to fit into grid
            if (window.innerWidth <= Util.SmallMediaLimit) {
                this.smallMedia = true
            }

            console.debug("Waiting for load", this.images)
            await Util.Images.LoadImages(this.images, this.entries.length, urlFunc, idFunc, true)
            console.debug("Loaded", this.images.size)

            if (this.postInit !== undefined) {
                await this.postInit()
            }
        },

        ...extra
    }
}