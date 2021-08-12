export const name = 'mangadex-manager';

import * as API from "/static/js/api.js"
import * as Util from "/static/js/util.js"

export default function () {
    return {
        source: {},
        downloads: [],
        paused: false,
        loaded: false,

        fmtPercent(p) {
            return Util.Fmt.Percent(p)
        },

        async deleteFinished() {
          await API.Download.DeleteFinishedTasks()
        },

        async retryFailed() {
            await API.Download.RetryFailedTasks()
        },

        async pauseDl() {
            await API.Download.Pause()
        },

        async resumeDl() {
            await API.Download.Resume()
        },

        async cancelDl() {
            await API.Download.Cancel()
        },

        async init() {
            this.source = API.Download.Manager()

            this.source.addEventListener('message', (e) => {
                this.loaded = true
                let data = JSON.parse(e.data)
                this.downloads = data.downloads
                this.paused = data.paused
            });
        }
    }
}