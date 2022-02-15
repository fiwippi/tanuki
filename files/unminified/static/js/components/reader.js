export const name = 'series';

import * as API from "/static/js/api.js"
import * as Util from "/static/js/util.js"

const modeList = [
    {name: 'Continuous', value: 'continuous'},
    {name: 'Paged (Right-to-Left)', value: 'paged-rl'},
    {name: 'Paged (Left-to-Right)', value: 'paged-lr'}
]

// https://stackoverflow.com/a/54000580
function isEventInElement(event, element)   {
    let rect = element.getBoundingClientRect();
    let x = event.clientX;
    if (x < rect.left || x >= rect.right) return false;
    let y = event.clientY;
    if (y < rect.top || y >= rect.bottom) return false;
    return true;
}

export default function (sid, eid, entry, initialProgress, entries, modal) {
    console.debug("entry progress:", initialProgress)

    return {
        // Reader data
        sid: sid,
        eid: eid,
        images: [],
        entry: entry,
        modal: modal,
        initialProgress: initialProgress,
        mode: localStorage.getItem('mode') || 'continuous',
        blankImage: Util.BlankImage,

        // Progress Data
        currentPage: 1,
        lastSavedPage: 1,
        acceptUpdates: false,

        // Modal Data
        modalAllowed: false,
        selectedPage: 1,
        selectedMode: "mode",
        selectedEntry: "",
        entryIndex: -1,
        entries: entries,
        hasNextEntry: false,
        hasPreviousEntry: false,

        get filteredEntries() {
            // Remove the current entry
            let temp = this.entries.filter(i => i.hash !== this.eid)
            // Prepend the entry so users can't change to it,
            // this works because in a select box, users can't
            // change to the first element
            let front = this.entries.filter(i => i.hash === this.eid)
            temp.unshift(front[0])
            return temp
        },

        get filteredPages() {
            let temp = Array.from({length: this.entry.pages}, (_, i) => i + 1).filter(i => i !== this.currentPage)
            temp.unshift(this.currentPage)
            return temp
        },

        get filteredMode() {
            let temp = modeList.filter(i => i.value !== this.mode)
            let first = modeList.filter(i => i.value === this.mode)
            temp.unshift(first[0])
            return temp
        },

        // Util functions

        fmtPercent(p) {
            return Util.Fmt.Percent(p)
        },

        // Webtoon functions
        webtoon: false,

        loadWebtoon() {
            let val = localStorage.getItem('webtoon')
            if (['true', 'false'].indexOf(val) >= 0) {
                this.webtoon = (val === 'true') ? true: false
            }
        },

        saveWebtoon() {
            if (this.webtoon === true) {
                localStorage.setItem('webtoon', 'true')
            } else {
                localStorage.setItem('webtoon', 'false')
            }
        },

        // Init related

        async init() {
            // Load the webtoon state
            this.loadWebtoon()

            // Redefine modal functionality, this stops bugs, e.g. if you are on
            // paged view and click to the left of the modal, it stops the page
            // trying to scroll left
            this.modal.show = () => {
                if (this.modalAllowed) {
                    this.modal.visible = true
                }
            }

            this.modal.hide = () => {
                this.modalAllowed = false
                this.modal.visible = false
                this.$nextTick(() => {
                    this.modalAllowed = true
                })
            }

            // Start at the specific page if specified, otherwise start at the user's last progress
            let specifiedPage = new URLSearchParams(window.location.search).get('page')
            if (specifiedPage === null || specifiedPage === undefined) {
                this.currentPage = this.initialProgress.current
            } else {
                this.currentPage = Number(specifiedPage)
            }

            // Current page could be zero if the initalProgress.current
            // or the specified page is zero so we max to ensure it's 1
            this.currentPage = Math.max(this.currentPage, 1)

            this.lastSavedPage = this.currentPage

            // Determine which buttons can be displayed
            this.entryIndex = this.entries.findIndex((e) => e.hash === this.eid)
            if (this.entryIndex + 1 < this.entries.length) {
                this.hasNextEntry = true
            }
            if (this.entryIndex - 1 >= 0) {
                this.hasPreviousEntry = true
            }

            // Initialise the images array
            this.images = new Array(this.entry.pages)

            // Ensure the mode is saved, jumpToPage triggers here
            this.changeMode(this.mode)

            // If we don't do this x-intersect triggers a new
            // update progress event before we can navigate to
            // the current page
            this.acceptUpdates = true
        },

        // Page functions

        async getPage(num, buffer) {
            // Ensure valid page num
            if (num < 1 || num > this.entry.pages) {
                console.debug("page out of reach", num)
                return
            }

            // Don't fetch page if already loaded
            if (this.images[num - 1] === undefined) {
                // Fetch the image
                let img = new Image()
                let p =  Util.Images.WaitForLoad(img)
                img.src = `/api/series/${this.sid}/entries/${this.eid}/page/${num}`
                let success = await p
                    .then(() => { return true })
                    .catch(err => {console.log("failed to load image", num, err); return false})
                // Set the img if loaded properly
                if (success) {
                    this.images[num - 1] = img
                }
            } else {
                console.debug("page already loaded", num)
            }

            // If scrolling we need to load the image onto the placeholder,
            // we do this outside the image fetching because there are cases
            // the image have already been fetched (e.g. with the paged reader)
            // but we haven't bound their src to the continuous pages src.
            // This way ensures binding always happens.
            if (this.mode === 'continuous') {
                document.getElementById(`page-${num}`).src = this.images[num - 1].src
                document.getElementById(`page-${num}`).classList.remove("unloaded")
            }

            // Regardless of whether the image is already loaded, attempt to buffer ahead and behind,
            // we only want the viewport changing to trigger buffering because otherwise we will
            // start recursion and load all images possible
            if (buffer) {
                // User could be scrolling up or down
                await this.getPage(num - 1, false)
                await this.getPage(num - 2, false)

                await this.getPage(num + 1, false)
                await this.getPage(num + 2, false)
                await this.getPage(num + 3, false)
            }
        },

        async jumpToPage(num) {
            // If called from the modal, num is a string,
            // so we need to ensure it is parsed as a number
            num = Number(num)
            // We might be changing because of the modal
            // and if so we don't want to see it
            this.modal.hide()
            // Don't want to jump to pages which are out of range
            if (num < 1 || num > this.entry.pages) {
                console.debug("page out of reach", num)
                return
            }

            if (this.mode === 'continuous') {
                document.getElementById(`page-${num}`).scrollIntoView(true);
            } else {
                // Manually call getPage since we can't use x-intersect
                await this.getPage(num, true)
                this.$nextTick(() => {
                    let page = document.getElementById("page")
                    page.src = this.images[num - 1].src
                    page.classList.remove("unloaded")
                })
            }

            // Checks whether we need to update the page progress
            await this.updateProgress(num)
        },

        // Modal functions

        showModal() {
            this.modal.show()
            this.$nextTick(() => {
                // Need to ensure the displayed element in the select boxes
                // are the selected ones, for some reason they aren't so
                // this fixes it
                let ps = document.getElementById("pageSelect")
                if (ps !== undefined) ps.selectedIndex = 0;

                let ms = document.getElementById("modeSelect")
                if (ms !== undefined) ms.selectedIndex = 0;
            })
        },

        changeMode(mode) {
            // Set the mode
            this.mode = mode
            localStorage.setItem('mode', this.mode)

            // Change styling according to the mode
            if (this.mode === 'continuous') {
                document.body.classList.remove('pagedBodyExtra')
                document.getElementById('container').classList.remove('pagedContainerExtra')
            } else {
                document.body.classList.add('pagedBodyExtra')
                document.getElementById('container').classList.add('pagedContainerExtra')
            }

            // Wait for alpine to update the page layout before jumping
            this.$nextTick(() => {
                this.jumpToPage(this.currentPage)
            })
        },

        changeEntry(eid) {
            let url = `/reader/${this.sid}/${eid}`
            window.location.replace(url)
        },

        exitReader() {
            let url = `/entries/${this.sid}`
            window.location.replace(url)
        },

        async handlePagedClick(e) {
            if (this.modal.visible || !this.modalAllowed) {
                return
            }

            let mouseX = e.clientX
            let width = document.body.clientWidth
            let percent = mouseX / width

            // Click in the left 40% means go left
            if (percent < 0.4) {
                await this.flipPage("left")
            }

            // Click in the centre 20% and click on the page means show the modal
            let page = document.getElementById("page")
            if (isEventInElement(e, page) && percent >= 0.4 && percent <= 0.6) {
                this.showModal()
            }

            // Click in the right 40% means go right
            if (percent > 0.6) {
                await this.flipPage("right")
            }
        },

        // Progress functions

        async updateProgress(num) {
            if (!this.acceptUpdates) {
                return
            }

            this.currentPage = num

            let a = this.currentPage === 1
            let b = this.currentPage === this.entry.pages
            let c = Math.abs(this.lastSavedPage - this.currentPage) >= 5

            if (a || b || c) {
                this.lastSavedPage = this.currentPage
                await API.Catalog.PatchProgress(this.sid, this.eid, this.currentPage.toString())
            }
        },

        // Paged reader functions

        async flipPage(direction) {
            if (!this.modalAllowed) {
                return
            }

            if (this.mode === 'paged-lr') {
                if (direction === 'left') await this.jumpToPage(this.currentPage - 1)
                if (direction === 'right') await this.jumpToPage(this.currentPage + 1)
            } else if (this.mode === 'paged-rl') {
                if (direction === 'left') await this.jumpToPage(this.currentPage + 1)
                if (direction === 'right') await this.jumpToPage(this.currentPage - 1)
            }
        },
    }
}