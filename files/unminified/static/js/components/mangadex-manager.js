export const name = 'mangadex-manager';

import * as API from "/static/js/api.js"
import * as Util from "/static/js/util.js"

export default function () {
    return {
        source: {},
        downloads: [],
        paused: false,
        waiting: false,
        waitingText: "Processing finished downloads (pauses until done)",
        loaded: false,
        loadingText: "Connecting to Server",

        fmtPercent(p) {
            return Util.Fmt.Percent(p)
        },

        async deleteAll() {
            await API.Download.DeleteAllDownloads()
                .catch(() => {alert("Failed!")})
        },
        async deleteSuccessful() {
            await API.Download.DeleteSuccessfulDownloads()
                .catch(() => {alert("Failed!")})
        },
        async retryFailed() {
            await API.Download.RetryFailedDownloads()
                .catch(() => {alert("Failed!")})
        },
        async pauseDl() {
            await API.Download.Pause()
                .catch(() => {alert("Failed!")})
        },
        async resumeDl() {
            await API.Download.Resume()
                .catch(() => {alert("Failed!")})
        },
        async cancelDl() {
            await API.Download.Cancel()
                .catch(() => {alert("Failed!")})
        },

        async init() {
            Util.Animate.DotDotDot("Processing finished downloads (pauses until done)", (str) => this.waitingText = str)
            let cancel = Util.Animate.DotDotDot("Connecting to Server", (str) => this.loadingText = str)
            this.source = API.Download.Manager()
            cancel()

            this.source.addEventListener('message', (e) => {
                this.loaded = true
                let data = JSON.parse(e.data)
                this.downloads = data.downloads
                this.paused = data.paused
                this.waiting = data.waiting
            });
        }
    }
}