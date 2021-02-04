function getEntries(url, method, data) {
    return axios({
        method: method,
        url: url,
        data: data,
        xsrfCookieName: 'csrftoken',
        xsrfHeaderName: 'X-CSRFToken',
        headers: {
            'X-Requested-With': 'XMLHttpRequest'
        }
    })
}

var app = new Vue({
    el: '#app',
    data: {
        keyword: '',
        entries: [
            { code: 'abc', name: 'ABC', open: 123, high: 456, low: 0, close: 200 },
            { code: 'abc', name: 'XYZ', open: 123, high: 456, low: 0, close: 200 }
        ]
    },
    created() {
        let r = getEntries('', 'get', '').then((res) => {
            this.entries = res.data.entries
        });
    },
    computed: {
        filteredEntries() {
            return this.entries.filter(entry => {
                entry.name.toLowerCase().match(this.keyword.toLowerCase())
            })
        }
    }
})