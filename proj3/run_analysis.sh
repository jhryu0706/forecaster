#!/bin/bash
#
#SBATCH --mail-user=jhryu@cs.uchicago.edu
#SBATCH --mail-type=ALL
#SBATCH --job-name=proj3_benchmark 
#SBATCH --chdir=./
#SBATCH --output=./slurm/out/%j.%N.stdout
#SBATCH --error=./slurm/out/%j.%N.stderr
#SBATCH --partition=general 
#SBATCH --nodes=1
#SBATCH --ntasks=1
#SBATCH --cpus-per-task=20
#SBATCH --mem-per-cpu=900
#SBATCH --exclusive
#SBATCH --time=4:00:00
#SBATCH --open-mode=append


echo "Fixing permissions..."

module load golang/1.19

go clean -modcache

echo "Fetching Go module dependencies..."
go get github.com/coocood/qbs@v0.0.0-20170418011607-8554e18a96c9
go get github.com/mattn/go-sqlite3@v1.14.22
go get gonum.org/v1/gonum@v0.15.0

# Indirect dependencies
go get github.com/coocood/mysql@v0.0.0-20130514171929-d22091ecccb5
go get github.com/lib/pq@v1.10.9
go get golang.org/x/exp@v0.0.0-20240506185415-9bf2ced13842

# Ensure dependencies are tidy and vendored
echo "Tidying and vendoring Go module dependencies..."
go mod tidy
go mod vendor

python3 benchmark/run.py

echo "Removing vendor directory..."
rm -rf vendor

echo "Job completed successfully."