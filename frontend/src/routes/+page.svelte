<script lang="ts">
    import { onMount } from "svelte";
    import Chart, { type TooltipItem } from "chart.js/auto";
    import {
        GetTransactions,
        AddTransaction,
        UpdateTransaction,
        DeleteTransaction,
        AddAccount,
        DeleteAccount,
        GetAccountBalances,
        GetTransfers,
        AddTransfer,
        DeleteTransfer,
        GetNote,
        SaveNote,
        AddInstallment,
        GetInstallments,
        DeleteInstallment,
        ExportJSON,
        ExportCSV,
        GetBackupConfig,
        SaveBackupConfig,
        BackupNow,
        RestoreFromBackup,
    } from "../../wailsjs/go/main/App";

    interface Transaction {
        id?: number;
        type: "expense" | "revenue";
        desc: string;
        cat: string;
        val: number;
        date: string;
        account_id?: number | null;
        installment_id?: number | null;
        installment_index?: number | null;
    }

    interface Installment {
        id?: number;
        desc: string;
        cat: string;
        total_val: number;
        n_installments: number;
        start_date: string;
        account_id?: number | null;
        paid_count?: number;
        monthly_val?: number;
    }

    interface Account {
        id?: number;
        name: string;
        initial_balance: number;
    }

    interface AccountBalance {
        id: number;
        name: string;
        balance: number;
    }

    interface Transfer {
        id?: number;
        from_account_id: number;
        to_account_id: number;
        from_account_name?: string;
        to_account_name?: string;
        amount: number;
        date: string;
        desc: string;
    }

    // ── State ────────────────────────────────────────────────
    let records = $state<Transaction[]>([]);
    let dateFrom = $state("");
    let dateTo = $state("");
    let activePeriod = $state("3m");

    // add-expense form
    let eVal = $state("");
    let eDesc = $state("");
    let eCat = $state("");
    let eDate = $state("");
    let eAccountId = $state<number | "">("");

    // add-revenue form
    let rVal = $state("");
    let rDesc = $state("");
    let rCat = $state("");
    let rDate = $state("");
    let rAccountId = $state<number | "">("");

    // edit form
    let editingId = $state<number | null>(null);
    let editType = $state<"expense" | "revenue">("expense");
    let editVal = $state("");
    let editDesc = $state("");
    let editCat = $state("");
    let editDate = $state("");
    let editAccountId = $state<number | "">("");

    // installments
    let installments = $state<Installment[]>([]);
    let eInstallMode = $state(false);
    let eInstallN = $state("");

    // dynamic category lists derived from all records ever loaded
    let expCategories = $derived(
        [
            ...new Set(
                records.filter((r) => r.type === "expense").map((r) => r.cat),
            ),
        ].sort(),
    );
    let revCategories = $derived(
        [
            ...new Set(
                records.filter((r) => r.type === "revenue").map((r) => r.cat),
            ),
        ].sort(),
    );

    // ── Derived period data ───────────────────────────────────
    let filtered = $derived(
        records.filter((r) => r.date >= dateFrom && r.date <= dateTo),
    );
    let expenses = $derived(filtered.filter((r) => r.type === "expense"));
    let revenues = $derived(filtered.filter((r) => r.type === "revenue"));
    let totalExp = $derived(expenses.reduce((s, r) => s + r.val, 0));
    let totalRev = $derived(revenues.reduce((s, r) => s + r.val, 0));
    let balance = $derived(totalRev - totalExp);
    let last4 = $derived([...records].sort((a, b) => (b.id ?? 0) - (a.id ?? 0)).slice(0, 4));
    let ledgerSorted = $derived(
        [...filtered].sort(
            (a, b) => b.date.localeCompare(a.date) || (b.id ?? 0) - (a.id ?? 0),
        ),
    );

    // ── Notes ─────────────────────────────────────────────────
    let expNote = $state("");
    let revNote = $state("");

    // ── Variation: compare current period to equally-sized previous period ──
    let varExp = $derived.by(() => {
        if (!dateFrom || !dateTo) return 0;
        const ms = new Date(dateTo).getTime() - new Date(dateFrom).getTime();
        const prevTo = new Date(new Date(dateFrom).getTime() - 1)
            .toISOString()
            .split("T")[0];
        const prevFrom = new Date(new Date(dateFrom).getTime() - ms - 1)
            .toISOString()
            .split("T")[0];
        const prevExp = records
            .filter(
                (r) =>
                    r.type === "expense" &&
                    r.date >= prevFrom &&
                    r.date <= prevTo,
            )
            .reduce((s, r) => s + r.val, 0);
        if (prevExp === 0) return null;
        return ((totalExp - prevExp) / prevExp) * 100;
    });

    let varRev = $derived.by(() => {
        if (!dateFrom || !dateTo) return 0;
        const ms = new Date(dateTo).getTime() - new Date(dateFrom).getTime();
        const prevTo = new Date(new Date(dateFrom).getTime() - 1)
            .toISOString()
            .split("T")[0];
        const prevFrom = new Date(new Date(dateFrom).getTime() - ms - 1)
            .toISOString()
            .split("T")[0];
        const prevRev = records
            .filter(
                (r) =>
                    r.type === "revenue" &&
                    r.date >= prevFrom &&
                    r.date <= prevTo,
            )
            .reduce((s, r) => s + r.val, 0);
        if (prevRev === 0) return null;
        return ((totalRev - prevRev) / prevRev) * 100;
    });

    // --- Exporting -----------------------
    async function exportJSON() {
        try {
            await ExportJSON();
        } catch (e) {
            console.error("export_json failed:", e);
        }
    }

    async function exportCSV() {
        try {
            await ExportCSV();
        } catch (e) {
            console.error("export_csv failed:", e);
        }
    }

    // ── Backup ────────────────────────────────────────────────
    interface BackupConfig {
        provider: string;
        host: string;
        repo: string;
        token: string;
    }

    let backupCfg = $state<BackupConfig>({
        provider: "github",
        host: "",
        repo: "",
        token: "",
    });
    let backupStatus = $state<string | null>(null);
    let backupBusy = $state(false);

    async function loadBackupConfig() {
        try {
            backupCfg = await GetBackupConfig();
        } catch (e) {
            console.error("get_backup_config failed:", e);
        }
    }

    async function saveBackupConfig() {
        try {
            await SaveBackupConfig(backupCfg);
            backupStatus = "saved";
        } catch (e) {
            backupStatus = String(e);
        }
    }

    async function doBackup() {
        backupBusy = true;
        backupStatus = "backing up...";
        try {
            await BackupNow();
            backupStatus = "backup complete";
        } catch (e) {
            backupStatus = String(e);
        } finally {
            backupBusy = false;
        }
    }

    async function doRestore() {
        if (
            !confirm(
                "This will replace ALL local data with the backup. Continue?",
            )
        )
            return;
        backupBusy = true;
        backupStatus = "restoring...";
        try {
            await RestoreFromBackup();
            backupStatus = "restore complete — reloading...";
            await loadAccounts();
            await loadTransfers();
            await loadInstallments();
            setPeriod(activePeriod || "3m");
            backupStatus = "restore complete";
        } catch (e) {
            backupStatus = String(e);
        } finally {
            backupBusy = false;
        }
    }

    // ── Accounts & Transfers ──────────────────────────────────
    let accountBalances = $state<AccountBalance[]>([]);
    let transfers = $state<Transfer[]>([]);
    let totalAccountBalance = $derived(
        accountBalances.reduce((s, a) => s + a.balance, 0),
    );

    // add-account form
    let aName = $state("");
    let aInitial = $state("");

    // transfer form
    let tfFrom = $state<number | "">("");
    let tfTo = $state<number | "">("");
    let tfAmount = $state("");
    let tfDesc = $state("");
    let tfDate = $state("");

    async function loadAccounts() {
        try {
            accountBalances = await GetAccountBalances();
        } catch (e) {
            console.error("get_account_balances failed:", e);
        }
    }

    async function loadTransfers() {
        try {
            transfers = await GetTransfers();
        } catch (e) {
            console.error("get_transfers failed:", e);
        }
    }

    async function addAccount() {
        const name = aName.trim();
        const initial_balance = parseFloat(aInitial) || 0;
        if (!name) return;
        try {
            await AddAccount({ name, initial_balance });
            aName = "";
            aInitial = "";
            await loadAccounts();
        } catch (e) {
            console.error("add_account failed:", e);
        }
    }

    async function deleteAccount(id: number) {
        try {
            await DeleteAccount(id);
            await loadAccounts();
        } catch (e) {
            console.error("delete_account failed:", e);
        }
    }

    async function addTransfer() {
        if (tfFrom === "" || tfTo === "" || tfFrom === tfTo) return;
        const amount = parseFloat(tfAmount);
        if (isNaN(amount) || amount <= 0 || !tfDate) return;
        try {
            await AddTransfer({
                from_account_id: tfFrom as number,
                to_account_id: tfTo as number,
                amount,
                date: tfDate,
                desc: tfDesc.trim(),
            });
            tfAmount = "";
            tfDesc = "";
            await loadAccounts();
            await loadTransfers();
        } catch (e) {
            console.error("add_transfer failed:", e);
        }
    }

    async function deleteTransfer(id: number) {
        try {
            await DeleteTransfer(id);
            await loadAccounts();
            await loadTransfers();
        } catch (e) {
            console.error("delete_transfer failed:", e);
        }
    }

    // ── Category breakdown ────────────────────────────────────
    function catBreakdown(items: Transaction[], total: number) {
        const m: Record<string, number> = {};
        items.forEach((r) => {
            m[r.cat] = (m[r.cat] || 0) + r.val;
        });
        return Object.entries(m)
            .sort((a, b) => b[1] - a[1])
            .map(([cat, val]) => ({
                cat,
                val,
                pct: total > 0 ? ((val / total) * 100).toFixed(1) : "0.0",
            }));
    }
    let expCats = $derived(catBreakdown(expenses, totalExp));
    let revCats = $derived(catBreakdown(revenues, totalRev));

    // ── Formatting ────────────────────────────────────────────
    const MO = [
        "Jan",
        "Feb",
        "Mar",
        "Apr",
        "May",
        "Jun",
        "Jul",
        "Aug",
        "Sep",
        "Oct",
        "Nov",
        "Dec",
    ];
    const brl = (v: number) =>
        "R$ " +
        v
            .toFixed(2)
            .replace(".", ",")
            .replace(/\B(?=(\d{3})+(?!\d))/g, ".");
    const fdt = (d: string) => {
        const [y, m, day] = d.split("-");
        return `${day} ${MO[+m - 1]} ${y}`;
    };
    const fmtVar = (v: number | null) =>
        v === null ? "n/a" : (v >= 0 ? "+" : "") + v.toFixed(1) + "%";

    // ── Period control ────────────────────────────────────────
    function setPeriod(p: string) {
        activePeriod = p;
        const now = new Date();
        const from = new Date(now);
        if (p === "1m") from.setMonth(from.getMonth() - 1);
        else if (p === "3m") from.setMonth(from.getMonth() - 3);
        else if (p === "6m") from.setMonth(from.getMonth() - 6);
        else if (p === "1y") from.setFullYear(from.getFullYear() - 1);
        dateFrom = from.toISOString().split("T")[0];
        dateTo = now.toISOString().split("T")[0];
        loadRecords();
        loadNotes();
    }

    function applyDates() {
        activePeriod = "";
        loadRecords();
        loadNotes();
    }

    // ── Backend calls ─────────────────────────────────────────
    async function loadRecords() {
        try {
            records = (await GetTransactions(dateFrom, dateTo)) as Transaction[];
        } catch (e) {
            console.error("get_transactions failed:", e);
        }
    }

    async function loadNotes() {
        try {
            expNote = await GetNote("expense", dateFrom, dateTo);
            revNote = await GetNote("revenue", dateFrom, dateTo);
        } catch (e) {
            console.error("get_note failed:", e);
        }
    }

    async function saveNote(section: "expense" | "revenue", content: string) {
        try {
            await SaveNote(section, dateFrom, dateTo, content);
        } catch (e) {
            console.error("save_note failed:", e);
        }
    }

    let accountMap = $derived(
        Object.fromEntries(
            accountBalances.map((a) => [a.id, a.name]),
        ) as Record<number, string>,
    );

    let installmentMap = $derived(
        Object.fromEntries(installments.map((i) => [i.id!, i])) as Record<
            number,
            Installment
        >,
    );

    async function addEntry(type: "expense" | "revenue") {
        const isExp = type === "expense";
        const val = parseFloat(isExp ? eVal : rVal);
        const desc = isExp ? eDesc.trim() : rDesc.trim();
        const cat = isExp ? eCat.trim() : rCat.trim();
        const date = isExp ? eDate : rDate;
        const account_id = isExp
            ? eAccountId === ""
                ? undefined
                : eAccountId
            : rAccountId === ""
              ? undefined
              : rAccountId;
        if (!desc || isNaN(val) || val <= 0 || !date || !cat) return;

        try {
            await AddTransaction({ type, desc, cat, val, date, account_id });
            if (isExp) {
                eVal = "";
                eDesc = "";
            } else {
                rVal = "";
                rDesc = "";
            }
            await loadRecords();
            await loadAccounts();
        } catch (e) {
            console.error("add_transaction failed:", e);
        }
    }

    async function loadInstallments() {
        try {
            installments = await GetInstallments();
        } catch (e) {
            console.error("get_installments failed:", e);
        }
    }

    async function addInstallment() {
        const val = parseFloat(eVal);
        const desc = eDesc.trim();
        const cat = eCat.trim();
        const n = parseInt(eInstallN);
        const account_id = eAccountId === "" ? undefined : eAccountId;
        if (
            !desc ||
            isNaN(val) ||
            val <= 0 ||
            !eDate ||
            !cat ||
            isNaN(n) ||
            n < 2
        )
            return;
        try {
            await AddInstallment({
                desc,
                cat,
                total_val: val,
                n_installments: n,
                start_date: eDate,
                account_id,
            });
            eVal = "";
            eDesc = "";
            eInstallN = "";
            eInstallMode = false;
            await loadRecords();
            await loadAccounts();
            await loadInstallments();
        } catch (e) {
            console.error("add_installment failed:", e);
        }
    }

    async function deleteInstallment(id: number) {
        try {
            await DeleteInstallment(id);
            await loadRecords();
            await loadAccounts();
            await loadInstallments();
        } catch (e) {
            console.error("delete_installment failed:", e);
        }
    }

    async function deleteEntry(id: number) {
        try {
            await DeleteTransaction(id);
            editingId = null;
            await loadRecords();
            await loadAccounts();
        } catch (e) {
            console.error("delete_transaction failed:", e);
        }
    }

    function startEdit(r: Transaction) {
        editingId = r.id ?? null;
        editType = r.type;
        editVal = String(r.val);
        editDesc = r.desc;
        editCat = r.cat;
        editDate = r.date;
        editAccountId = r.account_id ?? "";
    }

    function cancelEdit() {
        editingId = null;
    }

    async function saveEdit() {
        if (editingId === null) return;
        const val = parseFloat(editVal);
        if (
            !editDesc.trim() ||
            !editCat.trim() ||
            isNaN(val) ||
            val <= 0 ||
            !editDate
        )
            return;
        try {
            await UpdateTransaction({
                id: editingId,
                type: editType,
                desc: editDesc.trim(),
                cat: editCat.trim(),
                val,
                date: editDate,
                account_id: editAccountId === "" ? undefined : editAccountId,
            });
            editingId = null;
            await loadRecords();
            await loadAccounts();
        } catch (e) {
            console.error("update_transaction failed:", e);
        }
    }

    // ── Charts ────────────────────────────────────────────────
    const chartOpts = (_color?: string) => ({
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
            legend: {
                labels: {
                    font: { family: "'Caveat', cursive", size: 13 },
                    color: "#4a3c2e",
                    boxWidth: 10,
                },
            },
            tooltip: {
                backgroundColor: "#1c140d",
                titleColor: "#f4eed8",
                bodyColor: "#c8bda6",
                titleFont: { family: "'Caveat', cursive", size: 14 },
                bodyFont: { family: "'Caveat', cursive", size: 13 },
                callbacks: {
                    label: (c: TooltipItem<"line">) =>
                        " " +
                        (c.dataset.label ?? "") +
                        ": " +
                        brl(c.parsed.y ?? 0),
                },
            },
        },
        scales: {
            x: {
                ticks: {
                    font: { family: "'Caveat', cursive", size: 12 },
                    color: "#8a7a66",
                },
                grid: { color: "#e5ddc8" },
                border: { color: "#c8bda6" },
            },
            y: {
                ticks: {
                    font: { family: "'Caveat', cursive", size: 12 },
                    color: "#8a7a66",
                    callback: (v: number | string) =>
                        "R$" + Number(v).toLocaleString("pt-BR"),
                },
                grid: { color: "#e5ddc8" },
                border: { color: "#c8bda6" },
            },
        },
    });

    function bucketByDay(items: Transaction[]) {
        const b: Record<string, number> = {};
        items.forEach((r) => {
            b[r.date] = (b[r.date] || 0) + r.val;
        });
        const ks = Object.keys(b).sort();
        return {
            ks,
            labels: ks.map((k) => {
                const [, m, d] = k.split("-");
                return d + " " + MO[+m - 1];
            }),
            vals: ks.map((k) => b[k]),
        };
    }

    // canvas refs
    let lineEl: HTMLCanvasElement | undefined;
    let donutEl: HTMLCanvasElement | undefined;
    let expLineEl: HTMLCanvasElement | undefined;
    let revLineEl: HTMLCanvasElement | undefined;
    let lc: Chart | undefined;
    let dc: Chart | undefined;
    let elc: Chart | undefined;
    let rlc: Chart | undefined;

    $effect(() => {
        // main line chart
        const bm: Record<string, { e: number; r: number }> = {};
        filtered.forEach((r) => {
            if (!bm[r.date]) bm[r.date] = { e: 0, r: 0 };
            bm[r.date][r.type === "expense" ? "e" : "r"] += r.val;
        });
        const ks = Object.keys(bm).sort();
        const labels = ks.map((k) => {
            const [, m, d] = k.split("-");
            return d + " " + MO[+m - 1];
        });
        if (lc) lc.destroy();
        if (lineEl)
            lc = new Chart(lineEl, {
                type: "line",
                data: {
                    labels,
                    datasets: [
                        {
                            label: "Expenses",
                            data: ks.map((k) => bm[k]?.e || 0),
                            borderColor: "#8c1f1f",
                            backgroundColor: "rgba(140,31,31,.07)",
                            borderWidth: 1.5,
                            pointRadius: 2,
                            pointBackgroundColor: "#8c1f1f",
                            tension: 0.4,
                            fill: true,
                        },
                        {
                            label: "Revenue",
                            data: ks.map((k) => bm[k]?.r || 0),
                            borderColor: "#1a5c30",
                            backgroundColor: "rgba(26,92,48,.07)",
                            borderWidth: 1.5,
                            pointRadius: 2,
                            pointBackgroundColor: "#1a5c30",
                            tension: 0.4,
                            fill: true,
                        },
                    ],
                },
                options: chartOpts(),
            });

        // donut
        const pal = [
            "#8c1f1f",
            "#b84040",
            "#c97060",
            "#d49880",
            "#a05030",
            "#7a3010",
            "#5a2008",
        ];
        const em: Record<string, number> = {};
        expenses.forEach((r) => {
            em[r.cat] = (em[r.cat] || 0) + r.val;
        });
        if (dc) dc.destroy();
        if (donutEl)
            dc = new Chart(donutEl, {
                type: "doughnut",
                data: {
                    labels: Object.keys(em),
                    datasets: [
                        {
                            data: Object.values(em),
                            backgroundColor: pal.slice(
                                0,
                                Object.keys(em).length,
                            ),
                            borderColor: "#f4eed8",
                            borderWidth: 2,
                        },
                    ],
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    cutout: "76%",
                    plugins: {
                        legend: {
                            position: "right",
                            labels: {
                                font: { family: "'Caveat',cursive", size: 12 },
                                color: "#4a3c2e",
                                boxWidth: 10,
                                padding: 8,
                            },
                        },
                        tooltip: {
                            backgroundColor: "#1c140d",
                            titleColor: "#f4eed8",
                            bodyColor: "#c8bda6",
                            titleFont: { family: "'Caveat',cursive", size: 13 },
                            bodyFont: { family: "'Caveat',cursive", size: 12 },
                            callbacks: {
                                label: (c) => {
                                    const p =
                                        totalExp > 0
                                            ? (
                                                  (c.parsed / totalExp) *
                                                  100
                                              ).toFixed(1)
                                            : 0;
                                    return (
                                        " " + brl(c.parsed) + " (" + p + "%)"
                                    );
                                },
                            },
                        },
                    },
                },
            });

        // expenses line
        const eb = bucketByDay(expenses);
        if (elc) elc.destroy();
        if (expLineEl)
            elc = new Chart(expLineEl, {
                type: "line",
                data: {
                    labels: eb.labels,
                    datasets: [
                        {
                            label: "Expenses",
                            data: eb.vals,
                            borderColor: "#8c1f1f",
                            backgroundColor: "rgba(140,31,31,.07)",
                            borderWidth: 1.5,
                            pointRadius: 2,
                            pointBackgroundColor: "#8c1f1f",
                            tension: 0.4,
                            fill: true,
                        },
                    ],
                },
                options: chartOpts(),
            });

        // revenues line
        const rb = bucketByDay(revenues);
        if (rlc) rlc.destroy();
        if (revLineEl)
            rlc = new Chart(revLineEl, {
                type: "line",
                data: {
                    labels: rb.labels,
                    datasets: [
                        {
                            label: "Revenue",
                            data: rb.vals,
                            borderColor: "#1a5c30",
                            backgroundColor: "rgba(26,92,48,.07)",
                            borderWidth: 1.5,
                            pointRadius: 2,
                            pointBackgroundColor: "#1a5c30",
                            tension: 0.4,
                            fill: true,
                        },
                    ],
                },
                options: chartOpts(),
            });

        return () => {
            lc?.destroy();
            dc?.destroy();
            elc?.destroy();
            rlc?.destroy();
        };
    });

    // ── Init ──────────────────────────────────────────────────
    onMount(() => {
        const today = new Date().toISOString().split("T")[0];
        eDate = today;
        rDate = today;
        tfDate = today;
        setPeriod("3m");
        loadAccounts();
        loadTransfers();
        loadInstallments();
        loadBackupConfig();
    });
</script>

<!-- ══════════════════════════════════════════════════════════ TOPBAR -->
<header id="topbar">
    <div class="brand">Self Ledger</div>

    <div class="period-ctrl">
        {#each [["1m", "1m"], ["3m", "3m"], ["6m", "6m"], ["1y", "1y"]] as [p, label]}
            <button
                class="pbtn"
                class:on={activePeriod === p}
                onclick={() => setPeriod(p)}>{label}</button
            >
        {/each}
        <span class="psep">|</span>
        <input
            type="date"
            class="dinput"
            bind:value={dateFrom}
            onchange={applyDates}
        />
        <span style="color:var(--ink-faint);padding:0 4px">→</span>
        <input
            type="date"
            class="dinput"
            bind:value={dateTo}
            onchange={applyDates}
        />
    </div>

    <div class="nav-right">
        <div class="nav-totals">
            <div class="nav-tot">
                <span class="nav-tot-label">Expenses</span>
                <span class="nav-tot-val exp">{brl(totalExp)}</span>
            </div>
            <div class="nav-tot">
                <span class="nav-tot-label">Revenues</span>
                <span class="nav-tot-val rev">{brl(totalRev)}</span>
            </div>
            <div class="nav-tot">
                <span class="nav-tot-label">Balance</span>
                <span
                    class="nav-tot-val"
                    class:pos={totalAccountBalance >= 0}
                    class:neg={totalAccountBalance < 0}
                    >{brl(Math.abs(totalAccountBalance))}</span
                >
            </div>
        </div>

        <div class="nav-exports">
            <button class="exp-btn" onclick={exportJSON}>↓ JSON</button>
            <button class="exp-btn" onclick={exportCSV}>↓ CSV</button>
        </div>
    </div>
</header>

<!-- ══════════════════════════════════════════════════════════ PAGE -->
<div id="wrap">
    <div id="inner">
        <!-- ══ SECTION 1 — Last 4 records | Line chart ══ -->
        <div class="section wide-left">
            <div class="vcol">
                <div>
                    {#each last4 as r}
                        <div class="row row-3">
                            <span
                                class="val"
                                class:exp={r.type === "expense"}
                                class:rev={r.type === "revenue"}
                                >{brl(r.val)}</span
                            >
                            <span class="meta">{r.desc} · {r.cat}</span>
                            <span class="dt">{fdt(r.date)}</span>
                        </div>
                    {/each}
                </div>

                <div style="height:20px"></div>

                <!-- ADD EXPENSE -->
                <div class="add-row">
                    <input
                        class="ii v exp-ii"
                        type="number"
                        placeholder="R$ 0,00"
                        step=".01"
                        min="0"
                        bind:value={eVal}
                    />
                    <input
                        class="ii d"
                        type="text"
                        placeholder="description"
                        bind:value={eDesc}
                    />
                    <div class="cat-combo">
                        <input
                            class="ii cat-ii exp-ii"
                            type="text"
                            placeholder="category"
                            list="exp-cats-list"
                            bind:value={eCat}
                        />
                        <datalist id="exp-cats-list">
                            {#each expCategories as c}<option value={c}></option>{/each}
                        </datalist>
                    </div>
                    <input class="ii t" type="date" bind:value={eDate} />
                    <select class="ii tf-sel" bind:value={eAccountId}>
                        <option value="">account</option>
                        {#each accountBalances as a}
                            <option value={a.id}>{a.name}</option>
                        {/each}
                    </select>
                    {#if eInstallMode}
                        <input
                            class="ii install-n-ii exp-ii"
                            type="number"
                            placeholder="× n"
                            min="2"
                            step="1"
                            bind:value={eInstallN}
                        />
                        <button
                            class="abtn exp-abtn"
                            onclick={addInstallment}
                            title="add as installments">÷↵</button
                        >
                    {:else}
                        <button
                            class="abtn exp-abtn"
                            onclick={() => addEntry("expense")}>↵</button
                        >
                    {/if}
                    <button
                        class="abtn install-toggle"
                        class:active={eInstallMode}
                        onclick={() => {
                            eInstallMode = !eInstallMode;
                            eInstallN = "";
                        }}
                        title="split into installments">÷</button
                    >
                </div>

                <!-- ADD REVENUE -->
                <div class="add-row">
                    <input
                        class="ii v rev-ii"
                        type="number"
                        placeholder="R$ 0,00"
                        step=".01"
                        min="0"
                        bind:value={rVal}
                    />
                    <input
                        class="ii d"
                        type="text"
                        placeholder="description"
                        bind:value={rDesc}
                    />
                    <div class="cat-combo">
                        <input
                            class="ii cat-ii rev-ii"
                            type="text"
                            placeholder="category"
                            list="rev-cats-list"
                            bind:value={rCat}
                        />
                        <datalist id="rev-cats-list">
                            {#each revCategories as c}<option value={c}></option>{/each}
                        </datalist>
                    </div>
                    <input class="ii t" type="date" bind:value={rDate} />
                    <select class="ii tf-sel" bind:value={rAccountId}>
                        <option value="">account</option>
                        {#each accountBalances as a}
                            <option value={a.id}>{a.name}</option>
                        {/each}
                    </select>
                    <button
                        class="abtn rev-abtn"
                        onclick={() => addEntry("revenue")}>↵</button
                    >
                </div>
            </div>
            <div class="vdiv"></div>
            <div class="vcol">
                <div class="chart-wrap h200">
                    <canvas bind:this={lineEl}></canvas>
                </div>
            </div>
        </div>

        <!-- ══ SECTION 2 — Donut | Expense category list ══ -->
        <div class="section">
            <div class="vcol">
                <div class="chart-label">expenses per category</div>
                <div class="chart-wrap h160">
                    <canvas bind:this={donutEl}></canvas>
                </div>
            </div>
            <div class="vdiv"></div>
            <div class="vcol">
                {#each expCats as { cat, pct }}
                    <div class="cat-row">
                        <span class="cat-name">{cat}</span>
                        <span class="cat-pct exp">{pct}%</span>
                    </div>
                {/each}
            </div>
        </div>

        <!-- ══ SECTION 3 — Expenses ══ -->
        <div class="block-section">
            <div class="block-title exp">expenses</div>
            <div class="block-top">
                <div>
                    <div class="chart-wrap h200">
                        <canvas bind:this={expLineEl}></canvas>
                    </div>
                </div>
                <div class="annot" style="margin-top:0;">
                    <div class="annot-label">notes</div>
                    <textarea
                        rows="8"
                        placeholder="personal annotations..."
                        bind:value={expNote}
                        onblur={() => saveNote("expense", expNote)}
                    ></textarea>
                </div>
            </div>
            <div class="block-cats">
                <div class="sec-label" style="margin-bottom:8px;">
                    by category
                </div>
                <div class="block-cats-grid">
                    {#each expCats as { cat, pct }}
                        <div class="cat-row">
                            <span class="cat-name">{cat}</span>
                            <span class="cat-pct exp">{pct}%</span>
                        </div>
                    {/each}
                </div>
            </div>
        </div>

        <!-- ══ SECTION 4 — Revenues ══ -->
        <div class="block-section">
            <div class="block-title rev">revenues</div>
            <div class="block-top">
                <div>
                    <div class="chart-wrap h200">
                        <canvas bind:this={revLineEl}></canvas>
                    </div>
                </div>
                <div class="annot" style="margin-top:0;">
                    <div class="annot-label">notes</div>
                    <textarea
                        rows="8"
                        placeholder="personal annotations..."
                        bind:value={revNote}
                        onblur={() => saveNote("revenue", revNote)}
                    ></textarea>
                </div>
            </div>
            <div class="block-cats">
                <div class="sec-label" style="margin-bottom:8px;">
                    by category
                </div>
                <div class="block-cats-grid">
                    {#each revCats as { cat, pct }}
                        <div class="cat-row">
                            <span class="cat-name">{cat}</span>
                            <span class="cat-pct rev">{pct}%</span>
                        </div>
                    {/each}
                </div>
            </div>
        </div>

        <!-- ══ SECTION 5 — Full ledger | Variation ══ -->
        <div class="section wide-left">
            <div class="vcol">
                <div class="drule"></div>
                <div class="sec-label" style="margin-top:10px;">ledger</div>
                {#each ledgerSorted as r}
                    {#if editingId === r.id}
                        <div class="edit-row">
                            <select
                                class="ii tf-sel edit-type-sel"
                                bind:value={editType}
                            >
                                <option value="expense">expense</option>
                                <option value="revenue">revenue</option>
                            </select>
                            <input
                                class="ii v"
                                type="number"
                                step=".01"
                                min="0"
                                bind:value={editVal}
                            />
                            <input
                                class="ii d"
                                type="text"
                                placeholder="description"
                                bind:value={editDesc}
                            />
                            <input
                                class="ii cat-ii"
                                type="text"
                                placeholder="category"
                                bind:value={editCat}
                            />
                            <input
                                class="ii t"
                                type="date"
                                bind:value={editDate}
                            />
                            <select
                                class="ii tf-sel"
                                bind:value={editAccountId}
                            >
                                <option value="">no account</option>
                                {#each accountBalances as a}
                                    <option value={a.id}>{a.name}</option>
                                {/each}
                            </select>
                            <button class="abtn rev-abtn" onclick={saveEdit}
                                >✓</button
                            >
                            <button
                                class="abtn"
                                style="background:none;color:var(--ink-faint);"
                                onclick={cancelEdit}>✕</button
                            >
                            <button
                                class="acct-del edit-del-btn"
                                onclick={() => deleteEntry(editingId!)}
                                title="delete record">×</button
                            >
                        </div>
                    {:else}
                        <div class="row row-ledger">
                            <span
                                class="val"
                                class:exp={r.type === "expense"}
                                class:rev={r.type === "revenue"}
                                >{brl(r.val)}</span
                            >
                            <span class="meta"
                                >{r.desc} · {r.cat}{r.installment_id &&
                                installmentMap[r.installment_id]
                                    ? ` (${r.installment_index}/${installmentMap[r.installment_id].n_installments})`
                                    : ""}{r.account_id
                                    ? " · " + accountMap[r.account_id]
                                    : ""}</span
                            >
                            <span class="dt">{fdt(r.date)}</span>
                            <button
                                class="row-edit-btn"
                                onclick={() => startEdit(r)}
                                title="edit">✎</button
                            >
                        </div>
                    {/if}
                {/each}
            </div>
            <div class="vdiv"></div>
            <div
                class="vcol"
                style="display:flex;flex-direction:column;justify-content:center;"
            >
                <div class="var-row">
                    <span class="var-letter rev">R</span>
                    <span
                        class="var-val"
                        class:up={varRev !== null && varRev >= 0}
                        class:dn={varRev !== null && varRev < 0}
                        >{fmtVar(varRev)}</span
                    >
                </div>
                <div class="var-row" style="margin-top:8px;">
                    <span class="var-letter exp">E</span>
                    <span
                        class="var-val"
                        class:dn={varExp !== null && varExp > 0}
                        class:up={varExp !== null && varExp <= 0}
                        >{fmtVar(varExp)}</span
                    >
                </div>
            </div>
        </div>
        <!-- ══ SECTION 6 — Accounts & Transfers ══ -->
        <div class="block-section">
            <div class="block-title acct">accounts</div>

            <div class="acct-body">
                <!-- Account cards -->
                <div class="acct-cards">
                    {#each accountBalances as a}
                        <div class="acct-card">
                            <span class="acct-card-name">{a.name}</span>
                            <span
                                class="acct-card-bal"
                                class:pos={a.balance >= 0}
                                class:neg={a.balance < 0}>{brl(a.balance)}</span
                            >
                            <button
                                class="acct-del"
                                onclick={() => deleteAccount(a.id)}
                                title="delete account">×</button
                            >
                        </div>
                    {/each}

                    <!-- Add account inline -->
                    <div class="acct-add-row">
                        <input
                            class="ii d"
                            type="text"
                            placeholder="account name"
                            bind:value={aName}
                            onkeydown={(e) => e.key === "Enter" && addAccount()}
                        />
                        <input
                            class="ii v"
                            type="number"
                            placeholder="initial balance"
                            step=".01"
                            bind:value={aInitial}
                            onkeydown={(e) => e.key === "Enter" && addAccount()}
                        />
                        <button class="abtn" onclick={addAccount}>↵</button>
                    </div>
                </div>

                <div class="vdiv"></div>

                <!-- Transfer form + list -->
                <div class="vcol">
                    <div class="sec-label" style="margin-bottom:10px;">
                        record transfer
                    </div>
                    <div class="tf-form">
                        <select class="ii tf-sel" bind:value={tfFrom}>
                            <option value="" disabled>from</option>
                            {#each accountBalances as a}
                                <option value={a.id}>{a.name}</option>
                            {/each}
                        </select>
                        <span class="tf-arrow">→</span>
                        <select class="ii tf-sel" bind:value={tfTo}>
                            <option value="" disabled>to</option>
                            {#each accountBalances as a}
                                <option value={a.id}>{a.name}</option>
                            {/each}
                        </select>
                        <input
                            class="ii v"
                            type="number"
                            placeholder="R$ 0,00"
                            step=".01"
                            min="0"
                            bind:value={tfAmount}
                        />
                        <input
                            class="ii d"
                            type="text"
                            placeholder="description"
                            bind:value={tfDesc}
                        />
                        <input class="ii t" type="date" bind:value={tfDate} />
                        <button class="abtn" onclick={addTransfer}>↵</button>
                    </div>

                    <div class="drule" style="margin:12px 0 8px;"></div>

                    {#each transfers as t}
                        <div class="row tf-row">
                            <span class="tf-route">
                                {t.from_account_name} → {t.to_account_name}
                            </span>
                            <span class="val pos">{brl(t.amount)}</span>
                            <span class="meta"
                                >{t.desc ? t.desc + " · " : ""}{fdt(
                                    t.date,
                                )}</span
                            >
                            <button
                                class="acct-del"
                                onclick={() => deleteTransfer(t.id!)}
                                title="delete transfer">×</button
                            >
                        </div>
                    {/each}
                </div>
            </div>
        </div>
        <!-- ══ SECTION 7 — Installments ══ -->
        {#if installments.length > 0}
            <div class="block-section">
                <div class="block-title inst">installments</div>
                <div class="inst-list">
                    {#each installments as inst}
                        <div class="inst-row">
                            <div class="inst-info">
                                <span class="inst-desc">{inst.desc}</span>
                                <span class="inst-cat">{inst.cat}</span>
                                {#if inst.account_id && accountMap[inst.account_id]}
                                    <span class="inst-acct"
                                        >{accountMap[inst.account_id]}</span
                                    >
                                {/if}
                            </div>
                            <div class="inst-numbers">
                                <span class="inst-monthly exp"
                                    >{brl(inst.monthly_val ?? 0)}</span
                                >
                                <span class="inst-x">×</span>
                                <span class="inst-n"
                                    >{inst.n_installments}×</span
                                >
                                <span class="inst-total"
                                    >{brl(inst.total_val)}</span
                                >
                                <span class="inst-from"
                                    >{fdt(inst.start_date)}</span
                                >
                                <span class="inst-progress"
                                    >({inst.paid_count}/{inst.n_installments})</span
                                >
                            </div>
                            <button
                                class="acct-del"
                                onclick={() => deleteInstallment(inst.id!)}
                                title="delete installment plan">×</button
                            >
                        </div>
                    {/each}
                </div>
            </div>
        {/if}

        <!-- ══ SECTION — Git Backup ══ -->
        <div class="block-section">
            <div class="block-title" style="color:var(--ink-faint)">
                git backup
            </div>
            <div
                style="display:flex;gap:10px;align-items:baseline;flex-wrap:wrap;margin-top:10px;"
            >
                <select class="ii" bind:value={backupCfg.provider}>
                    <option value="github">GitHub</option>
                    <option value="gitlab">GitLab</option>
                    <option value="forgejo">Forgejo</option>
                    <option value="gitea">Gitea</option>
                    <option value="custom">Custom</option>
                </select>
                {#if backupCfg.provider !== "github" && backupCfg.provider !== "gitlab"}
                    <input
                        class="ii"
                        type="text"
                        placeholder="host (e.g. codeberg.org)"
                        bind:value={backupCfg.host}
                    />
                {/if}
                <input
                    class="ii"
                    type="text"
                    placeholder="owner/repo"
                    bind:value={backupCfg.repo}
                />
                <input
                    class="ii"
                    type="password"
                    placeholder="token"
                    bind:value={backupCfg.token}
                />
                <button class="abtn" onclick={saveBackupConfig}>save</button>
                <button
                    class="exp-btn"
                    onclick={doBackup}
                    disabled={backupBusy || !backupCfg.repo}>backup now</button
                >
                <button
                    class="exp-btn"
                    onclick={doRestore}
                    disabled={backupBusy || !backupCfg.repo}>restore</button
                >
            </div>
            {#if backupStatus}
                <div
                    style="font-size:13px;color:var(--ink-faint);margin-top:6px"
                >
                    {backupStatus}
                </div>
            {/if}
        </div>
    </div>
</div>

<style>
    :global(:root) {
        --paper: #f4eed8;
        --ink: #1c140d;
        --ink-mid: #4a3c2e;
        --ink-faint: #8a7a66;
        --rule: #c8bda6;
        --rule-dark: #9a8a74;
        --red: #8c1f1f;
        --green: #1a5c30;
    }
    :global(*) {
        box-sizing: border-box;
        margin: 0;
        padding: 0;
    }
    :global(body) {
        background: var(--paper);
        color: var(--ink);
        font-family: "Caveat", cursive;
        font-size: 17px;
        line-height: 1.5;
    }
    :global(body::after) {
        content: "";
        position: fixed;
        inset: 0;
        background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='300' height='300'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.7' numOctaves='4' stitchTiles='stitch'/%3E%3CfeColorMatrix type='saturate' values='0'/%3E%3C/filter%3E%3Crect width='300' height='300' filter='url(%23n)' opacity='0.07'/%3E%3C/svg%3E");
        pointer-events: none;
        z-index: 999;
    }

    /* ── TOPBAR ── */
    #topbar {
        position: fixed;
        top: 0;
        left: 0;
        right: 0;
        z-index: 100;
        background: var(--paper);
        padding: 8px 48px;
        display: grid;
        grid-template-columns: 1fr auto 1fr;
        align-items: center;
        border-bottom: 2.5px solid var(--ink);
        box-shadow: 0 4.5px 0 rgba(28, 20, 13, 0.16);
        min-height: 64px;
    }
    .brand {
        font-family: "Caveat Brush", cursive;
        font-size: 24px;
        color: var(--ink);
    }
    .period-ctrl {
        display: flex;
        align-items: baseline;
        gap: 4px;
        justify-content: center;
    }
    .pbtn {
        background: none;
        border: none;
        border-bottom: 2px solid transparent;
        font-family: "Caveat", cursive;
        font-size: 16px;
        color: var(--ink-faint);
        cursor: pointer;
        padding: 2px 8px;
        transition: all 0.15s;
    }
    .pbtn:hover {
        color: var(--ink);
    }
    .pbtn.on {
        color: var(--ink);
        border-bottom-color: var(--ink);
    }
    .psep {
        color: var(--rule);
        padding: 0 6px;
    }
    .dinput {
        background: none;
        border: none;
        border-bottom: 1px solid var(--rule-dark);
        font-family: "Caveat", cursive;
        font-size: 15px;
        color: var(--ink-mid);
        padding: 0 4px;
        width: 108px;
        outline: none;
    }
    .dinput:focus {
        border-bottom-color: var(--ink);
    }
    .nav-right {
        display: flex;
        align-items: center;
        gap: 20px;
        justify-self: end;
    }
    .nav-totals {
        display: flex;
        gap: 24px;
        align-items: center;
    }
    .nav-tot {
        display: flex;
        flex-direction: column;
        align-items: flex-end;
    }
    .nav-tot-label {
        font-family: "Caveat Brush", cursive;
        font-size: 10px;
        letter-spacing: 0.18em;
        text-transform: uppercase;
        color: var(--ink-faint);
        line-height: 1;
    }
    .nav-tot-val {
        font-size: 20px;
        font-weight: 700;
        line-height: 1.2;
    }
    .nav-tot-val.exp {
        color: var(--red);
    }
    .nav-tot-val.rev {
        color: var(--green);
    }
    .nav-tot-val.pos {
        color: var(--green);
    }
    .nav-tot-val.neg {
        color: var(--red);
    }
    .nav-exports {
        display: flex;
        flex-direction: column;
        align-items: flex-end;
        gap: 5px;
    }
    .exp-btn {
        font-family: "Caveat Brush", cursive;
        font-size: 13px;
        letter-spacing: 0.03em;
        background: none;
        border: 1.5px solid var(--ink-mid);
        color: var(--ink-mid);
        border-radius: 4px;
        padding: 2px 10px;
        cursor: pointer;
        line-height: 1.4;
        transition:
            background 0.15s,
            color 0.15s,
            border-color 0.15s;
        width: 76px;
        text-align: center;
    }
    .exp-btn:hover {
        background: var(--ink-mid);
        color: var(--paper);
        border-color: var(--ink-mid);
    }

    /* ── LAYOUT ── */
    #wrap {
        margin-top: 64px;
    }
    #inner {
        max-width: 1500px;
        margin: 0 auto;
        padding: 0 32px;
    }
    .section {
        display: grid;
        grid-template-columns: 1fr 2px 1fr;
    }
    .section.wide-left {
        grid-template-columns: 60fr 2px 40fr;
    }
    /* ── SECTION SEPARATORS (hand-drawn double-line) ── */
    .section ~ .section,
    .section ~ .block-section,
    .block-section ~ .section,
    .block-section ~ .block-section {
        border-top: 1.5px solid var(--rule-dark);
        box-shadow: inset 0 4px 0 0 rgba(154, 138, 116, 0.28);
    }
    .vcol {
        padding: 28px 24px;
    }
    .vdiv {
        width: 2px;
        background: linear-gradient(
            to bottom,
            transparent 0%,
            var(--rule-dark) 5%,
            var(--rule-dark) 95%,
            transparent 100%
        );
    }

    /* ── TITLES ── */
    .sec-label {
        font-family: "Caveat Brush", cursive;
        font-size: 13px;
        letter-spacing: 0.2em;
        text-transform: uppercase;
        color: var(--ink-faint);
        margin-bottom: 10px;
        border-left: 2.5px solid var(--rule-dark);
        padding-left: 7px;
    }
    .block-title {
        font-family: "Caveat Brush", cursive;
        font-size: 28px;
        padding-bottom: 8px;
        margin-bottom: 12px;
        position: relative;
    }
    .block-title::after {
        content: "";
        position: absolute;
        bottom: 2px;
        left: -1px;
        right: 5px;
        height: 2.5px;
        background: currentColor;
        border-radius: 2px;
        transform: rotate(-0.2deg);
        opacity: 0.9;
    }
    .block-title::before {
        content: "";
        position: absolute;
        bottom: -1px;
        left: 4px;
        right: 1px;
        height: 1px;
        background: currentColor;
        opacity: 0.28;
        transform: rotate(0.15deg);
    }
    .block-title.exp {
        color: var(--red);
    }
    .block-title.rev {
        color: var(--green);
    }
    .block-title.acct {
        color: var(--ink-mid);
    }

    /* ── ROWS ── */
    .row {
        display: grid;
        align-items: baseline;
        padding: 5px 0;
        border-bottom: 1px solid var(--rule);
        gap: 10px;
    }
    .row-3 {
        grid-template-columns: 120px 1fr auto;
    }
    .val {
        font-size: 19px;
        font-weight: 700;
    }
    .val.exp {
        color: var(--red);
    }
    .val.rev {
        color: var(--green);
    }
    .meta {
        font-size: 16px;
        color: var(--ink-mid);
    }
    .dt {
        font-size: 14px;
        color: var(--ink-faint);
        white-space: nowrap;
        text-align: right;
    }

    /* ── ADD ROWS ── */
    .add-row {
        display: flex;
        align-items: baseline;
        gap: 8px;
        padding: 6px 0;
        border-bottom: 1px solid var(--rule);
    }
    .ii {
        background: none;
        border: none;
        border-bottom: 1px dashed var(--rule);
        font-family: "Caveat", cursive;
        font-size: 16px;
        color: var(--ink);
        outline: none;
        padding: 0 4px;
    }
    .ii:focus {
        border-bottom-color: var(--ink-mid);
    }
    .ii.v {
        width: 88px;
    }
    .ii.d {
        flex: 1;
    }
    .ii.t {
        width: 110px;
    }
    .exp-ii {
        color: var(--red);
    }
    .exp-ii::placeholder {
        color: var(--red);
        opacity: 0.6;
    }
    .rev-ii {
        color: var(--green);
    }
    .rev-ii::placeholder {
        color: var(--green);
        opacity: 0.6;
    }

    /* ── COMBOBOX CATEGORY ── */
    .cat-combo {
        position: relative;
        width: 120px;
    }
    .cat-ii {
        width: 100%;
    }

    .abtn {
        background: none;
        border: none;
        font-family: "Caveat", cursive;
        font-size: 17px;
        cursor: pointer;
        transition: color 0.15s;
        padding: 0 4px;
    }
    .exp-abtn {
        color: var(--red);
        opacity: 0.7;
    }
    .exp-abtn:hover {
        opacity: 1;
    }
    .rev-abtn {
        color: var(--green);
        opacity: 0.7;
    }
    .rev-abtn:hover {
        opacity: 1;
    }

    /* ── BLOCK SECTIONS ── */
    .block-section {
        padding: 28px 24px;
    }
    .block-top {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 32px;
        margin-bottom: 20px;
    }
    .block-cats {
        border-top: 1px solid var(--rule);
        padding-top: 16px;
    }
    .block-cats-grid {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
        gap: 0 24px;
    }

    /* ── CATEGORY ROWS ── */
    .cat-row {
        display: grid;
        grid-template-columns: 1fr auto;
        align-items: baseline;
        padding: 4px 0;
        border-bottom: 1px solid var(--rule);
        gap: 10px;
    }
    .cat-name {
        font-size: 16px;
        color: var(--ink-mid);
    }
    .cat-pct {
        font-size: 17px;
        font-weight: 700;
    }
    .cat-pct.exp {
        color: var(--red);
    }
    .cat-pct.rev {
        color: var(--green);
    }

    /* ── CHARTS ── */
    .chart-wrap {
        width: 100%;
        position: relative;
    }
    .chart-wrap.h200 {
        height: 200px;
    }
    .chart-wrap.h160 {
        height: 160px;
    }
    .chart-label {
        font-family: "Caveat Brush", cursive;
        font-size: 14px;
        letter-spacing: 0.08em;
        color: var(--ink-faint);
        margin-bottom: 8px;
    }

    /* ── ANNOTATION ── */
    .annot {
        margin-top: 14px;
    }
    .annot-label {
        font-family: "Caveat Brush", cursive;
        font-size: 12px;
        letter-spacing: 0.15em;
        text-transform: uppercase;
        color: var(--ink-faint);
    }
    .annot textarea {
        display: block;
        width: 100%;
        background: none;
        border: none;
        border-bottom: 1px dashed var(--rule);
        font-family: "Caveat", cursive;
        font-size: 16px;
        font-style: italic;
        color: var(--ink-mid);
        resize: none;
        outline: none;
        line-height: 2;
        padding: 2px 0;
        margin-top: 2px;
    }

    /* ── VARIATION ── */
    .var-row {
        display: grid;
        grid-template-columns: 36px 1fr;
        align-items: baseline;
        padding: 6px 0;
        border-bottom: 1px solid var(--rule);
        gap: 8px;
    }
    .var-letter {
        font-family: "Caveat Brush", cursive;
        font-size: 26px;
    }
    .var-letter.exp {
        color: var(--red);
    }
    .var-letter.rev {
        color: var(--green);
    }
    .var-val {
        font-family: "Caveat", cursive;
        font-size: 38px;
        font-weight: 700;
    }
    .var-val.up {
        color: var(--green);
    }
    .var-val.dn {
        color: var(--red);
    }
    /* ── DOUBLE RULE ── */
    .drule {
        border: none;
        border-top: 1px solid var(--rule-dark);
        margin: 0 0 1px;
        position: relative;
    }
    .drule::after {
        content: "";
        display: block;
        border-top: 1px solid var(--rule-dark);
        margin-top: 3px;
    }

    /* ── ACCOUNTS ── */
    .acct-body {
        display: grid;
        grid-template-columns: auto 1px 1fr;
        gap: 0 24px;
        align-items: start;
    }
    .acct-cards {
        display: flex;
        flex-direction: column;
        gap: 8px;
        min-width: 260px;
    }
    .acct-card {
        display: flex;
        align-items: baseline;
        gap: 10px;
        padding: 8px 10px;
        border: 1.5px solid var(--rule-dark);
        border-radius: 4px;
        position: relative;
    }
    .acct-card-name {
        font-family: "Caveat Brush", cursive;
        font-size: 16px;
        color: var(--ink-mid);
        flex: 1;
    }
    .acct-card-bal {
        font-size: 20px;
        font-weight: 700;
    }
    .acct-card-bal.pos {
        color: var(--green);
    }
    .acct-card-bal.neg {
        color: var(--red);
    }
    .acct-del {
        background: none;
        border: none;
        color: var(--ink-faint);
        font-size: 18px;
        cursor: pointer;
        line-height: 1;
        padding: 0 2px;
        margin-left: 4px;
    }
    .acct-del:hover {
        color: var(--red);
    }
    .acct-add-row {
        display: flex;
        gap: 6px;
        align-items: center;
        margin-top: 4px;
    }
    .tf-form {
        display: flex;
        flex-wrap: wrap;
        gap: 6px;
        align-items: center;
    }
    .tf-sel {
        background: var(--paper);
        border: 1.5px solid var(--rule-dark);
        border-radius: 3px;
        color: var(--ink);
        font-family: "Caveat", cursive;
        font-size: 16px;
        padding: 3px 6px;
        cursor: pointer;
    }
    .tf-arrow {
        color: var(--ink-faint);
        font-size: 16px;
    }
    .tf-row {
        display: grid;
        grid-template-columns: 1fr auto auto auto;
        gap: 10px;
        align-items: baseline;
    }
    .tf-route {
        font-family: "Caveat Brush", cursive;
        font-size: 15px;
        color: var(--ink-mid);
    }
    .row-ledger {
        display: grid;
        grid-template-columns: auto 1fr auto auto;
        gap: 10px;
        align-items: baseline;
        padding: 2px 0;
    }
    .row-edit-btn {
        background: none;
        border: none;
        color: var(--ink-faint);
        font-size: 15px;
        cursor: pointer;
        padding: 0;
        line-height: 1;
        opacity: 0;
        transition: opacity 0.15s;
    }
    .row-ledger:hover .row-edit-btn {
        opacity: 1;
    }
    .edit-row {
        display: flex;
        flex-wrap: wrap;
        gap: 5px;
        align-items: center;
        padding: 6px 0;
        border-top: 1px dashed var(--rule);
        border-bottom: 1px dashed var(--rule);
        margin: 2px 0;
    }
    .edit-type-sel {
        font-size: 13px;
        padding: 2px 4px;
    }
    .edit-del-btn {
        margin-left: 8px;
    }

    /* ── INSTALLMENTS ── */
    .block-title.inst {
        color: var(--ink-mid);
    }
    .inst-list {
        display: flex;
        flex-direction: column;
        gap: 0;
    }
    .inst-row {
        display: flex;
        align-items: baseline;
        gap: 16px;
        padding: 6px 0;
        border-bottom: 1px solid var(--rule);
    }
    .inst-info {
        display: flex;
        align-items: baseline;
        gap: 8px;
        flex: 1;
    }
    .inst-desc {
        font-size: 17px;
        color: var(--ink);
    }
    .inst-cat {
        font-size: 14px;
        color: var(--ink-faint);
    }
    .inst-acct {
        font-size: 13px;
        color: var(--ink-faint);
        font-style: italic;
    }
    .inst-numbers {
        display: flex;
        align-items: baseline;
        gap: 6px;
    }
    .inst-monthly {
        font-size: 18px;
        font-weight: 700;
    }
    .inst-x {
        font-size: 14px;
        color: var(--ink-faint);
    }
    .inst-n {
        font-size: 16px;
        color: var(--ink-mid);
    }
    .inst-total {
        font-size: 15px;
        color: var(--ink-faint);
    }
    .inst-from {
        font-size: 13px;
        color: var(--ink-faint);
    }
    .inst-progress {
        font-size: 13px;
        color: var(--ink-faint);
        font-style: italic;
    }
    .install-n-ii {
        width: 52px;
    }
    .install-toggle {
        color: var(--ink-faint);
        font-size: 15px;
        opacity: 0.6;
        border: 1px solid transparent;
        border-radius: 3px;
        padding: 0 3px;
        transition: all 0.15s;
    }
    .install-toggle:hover {
        opacity: 1;
        color: var(--red);
    }
    .install-toggle.active {
        opacity: 1;
        color: var(--red);
        border-color: var(--red);
    }
</style>
