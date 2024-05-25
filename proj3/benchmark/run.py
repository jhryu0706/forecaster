import pandas as pd
import subprocess
from statistics import mean
import matplotlib.pyplot as plt
import matplotlib
matplotlib.use('Agg')
import os
import sys

print("this is current path: ", os.getcwd())
forecaster_path = './forecaster/'
benchmark_path = './benchmark/'

def main():
    """
    Global variables
    """
    filetypes = ['small','mixed','large']
    
    threadcount = [2,4,6,8,12]
    def get_sequential_data():
        """
        Code to run sequential version
        """
        print("Starting sequential processing")
        df = pd.DataFrame(index=filetypes, columns=["seq_time"])
        for f in filetypes:
            with open(forecaster_path+f+"_input.txt", "r") as file:
                input_data = file.read()    
            print("Now running ",f)
            allresults=[]
            for _ in range(5):
                result = subprocess.run(["go","run",forecaster_path+"forecaster.go","s"], 
                input = input_data,stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
                print(result.stderr)
                if result.returncode == 0:
                    allresults.append(float(result.stdout.strip()))
            df.at[f,'seq_time']=mean(allresults)
        print(df)
        df.to_csv(benchmark_path+"sequential_output.txt", index=True)
        return df
    
    def get_parallel_data(is_work_stealing:bool):
        """
        Code to run parallel version
        """
        df = pd.DataFrame(index=filetypes, columns=threadcount)
        for f in filetypes:
            with open(forecaster_path+f+"_input.txt", "r") as file:
                input_data = file.read()
            print("Now running ",f)
            for t in threadcount:
                allresults =[]
                for iteration in range(5):
                    if is_work_stealing:
                        result = subprocess.run(["go", "run", forecaster_path+"forecaster.go","p",str(t)],input = input_data, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
                    else:
                        result = subprocess.run(["go", "run", forecaster_path+"forecaster.go","p",str(t),"w"],input = input_data, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
                    if result.returncode == 0:
                        allresults.append(float(result.stdout.strip()))
                print(f"All results for filetype: {f} threadcount {t}",allresults)
                df.at[f, t]=mean(allresults)
        print(df)
        if is_work_stealing:
            df.to_csv(benchmark_path+"workstealing_parallel_output.txt", index=True)
        else:
            df.to_csv(benchmark_path+"non_workstealing_parallel_output.txt", index=True)
        return df
    
    def get_speedup_graph(version, df, seq_df):
        speedup_t = seq_df['seq_time'].div(df.T).T
        print(speedup_t)
        
        # create plot
        plt.figure(figsize=(10,5))

        #df1.columns is x-axis and each row is a line
        for label ,row in speedup_t.iterrows():
            plt.plot(df.columns,row,label=label)

        plt.xlabel('Number of Threads')
        plt.ylabel('Speedup')
        plt.title(f'{version} Speedup Graph')
        plt.legend(title='Filesize')
        plt.grid(True)
        plt.savefig(benchmark_path+f"{version}_speedup.png", format='png', dpi=300)
    seq_df = get_sequential_data()
    nonworkstealing_df = get_parallel_data(False)
    sys.stdout.flush()
    workstealing_df=get_parallel_data(True)
    sys.stdout.flush()
    get_speedup_graph("test2_Forecaster_Nonworkstealing", nonworkstealing_df,seq_df)
    get_speedup_graph("test2_Forecaster_Workstealing", workstealing_df,seq_df)
    return

if __name__=='__main__':
    main()
