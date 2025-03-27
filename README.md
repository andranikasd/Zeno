# Zeno
**Zeno** is a modern, lightweight, open-source AWS Cost Analytics engine built with Go and designed for seamless Grafana integration.
🔥 Love this direction — let’s pause and **clarify the core problem Zeno is solving**, so you can build with purpose, attract contributors, and clearly communicate value.


## 🧠 The Problem: What are we trying to solve?

### ❓ TL;DR
> **“Understanding and predicting AWS cloud costs is hard, fragmented, and opaque — especially at scale.”**

---

## 🧩 The Real-World Problems

### 💸 1. **AWS billing is complex and hard to interpret**
- CUR contains **millions of line items**, often deeply nested
- Default tools (Billing Console, Cost Explorer) are **slow, limited, or too high-level**
- Difficult to **attribute costs by team, tag, environment**, etc.

---

### 🧺 2. **Third-party tools are either closed-source, expensive, or bloated**
- CloudHealth, CloudCheckr, or even native AWS tools cost $$$
- Netflix Ice is dead and outdated
- Kubecost only helps for Kubernetes
- **FinOps teams often build DIY spreadsheets**

---

### 🔍 3. **Visibility and accountability are missing**
- Engineers don’t see what they spend
- Finance doesn’t understand usage
- **Nobody owns waste**

---

### 🔮 4. **Forecasting is an afterthought**
- No clear trends, budgets, or future projections
- Budgets are usually **reactive**, not predictive
- Forecasting tools don’t integrate easily with dashboards

---

## 🧠 The Zeno Opportunity

**Zeno** solves this by offering:

| Problem                            | Zeno's Solution                        |
|------------------------------------|----------------------------------------|
| CUR is complex                     | Ingests, flattens, and interprets CURs |
| No single-pane visibility          | Exposes a clean API + Grafana plugin   |
| No fine-grained attribution        | Supports filters: tags, services, teams|
| No budget forecasting              | Predictive models coming soon 🔮       |
| Cloud cost tools are $$$/closed    | Zeno is open-source and self-hosted    |

---

## 📣 Elevator Pitch

> **Zeno** is a lightweight, open-source platform that ingests AWS billing data (CUR), transforms it into clean, filterable insights, and visualizes it through Grafana — giving engineers and finance teams the clarity, control, and cost a