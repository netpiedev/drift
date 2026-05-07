#!/usr/bin/env node

import { runDrift } from "../core/runner.js";

const args = process.argv.slice(2);
const code = await runDrift(args);
process.exit(code);
