/**
 * @brief This program tests the collision rate on a single variable
 *        when using rtm with n threads
 *
 * @file testrtm.cc
 * @author Craig Hesling
 * @date 2018-04-01
 */
#include <stdio.h>
#include <stdlib.h>
#include <iostream>
#include <thread>
#include <chrono>
#include <vector>
#include <immintrin.h>

using namespace std;
using namespace std::chrono;

#define PERCENT(numerator,denominator) \
		(((long double)(long long)(numerator)) / ((long double)(long long)(denominator))) * ((long double)100)

#define DEFAULT_LOOP_ATTEMPTS 10000
// #define DEFAULT_LOOP_ATTEMPTS 100000000
// #define DEFAULT_NUM_THREADS 1
#define DEFAULT_NUM_THREADS 2


struct params {
	int tid;
	unsigned long long loops;
};

struct result {
	unsigned long long success;
	unsigned long long failure;
	unsigned          *statuses;
};

unsigned long long int counter = 0;

#define print_status_flag(tid,loopi,status,flag) \
			if (status & flag)                   \
				printf("[Thread %2d] loop %d - %s\n", tid, loopi, #flag)


void thread_func(params *data, result *res) {
	unsigned long long success = 0, failed = 0;
	params *par = (params *)data;
	res->statuses = new unsigned[par->loops];
	// int tid   = par->tid;
	int loops = par->loops;

	// printf("Hello from thread %d\n", tid);

	for (int i = 0; i < loops; i++) {
		unsigned status;
		if ((status = _xbegin()) == _XBEGIN_STARTED) {
            
            // central variable to increment
			counter++;

			_xend();
			// printf("TH%d - %Ld\n", tid, counter);
			success++;
		} else {
			// printf("TH%d - %d - Failed\n", tid, i);
			failed++;
            // save status to analyze later
            res->statuses[i] = status;
		}
	}

	res->success = success;
	res->failure = failed;

	// printf("TH%d - Success=%Lu | Failed=%Lu\n", tid, success, failed);
	// fflush(stdout);
}

int main(int argc, char *argv[]) {
	int num_threads = DEFAULT_NUM_THREADS;
	int loops = DEFAULT_LOOP_ATTEMPTS;

	if (argc > 1) {
		num_threads = atoi(argv[1]);
	}

	if (argc > 2) {
		loops = atoi(argv[2]);
	}

	printf("# Spawning %d threads to do %d loops\n", num_threads, loops);

	thread threads[num_threads];
	params    *par[num_threads];
	result    *res[num_threads];

	for (int tid = 0; tid < num_threads; tid++) {
		par[tid] = new params;
		par[tid]->tid = tid;
		par[tid]->loops = loops;
		res[tid] = new result;
		res[tid]->statuses = new unsigned[par[tid]->loops];
	}

	const unsigned long long total = num_threads*loops;
	unsigned long long       success = 0;
	unsigned long long       failure = 0;

    /* Time launching and joining DEFAULT_NUM_THREADS threads */
	auto start = high_resolution_clock::now();
	for (int tid = 0; tid < num_threads; tid++) {
		threads[tid] = thread(thread_func, par[tid], res[tid]);
	}
	for (int tid = 0; tid < num_threads; tid++) {
		threads[tid].join();
	}
	auto stop = high_resolution_clock::now();

    /* Display stats for each thread */
	for (int tid = 0; tid < num_threads; tid++) {
        /* Give quick summary of successes and failures */
		success += res[tid]->success;
		failure += res[tid]->failure;
		printf("TH%3d - Success=%Lu | Failed=%Lu | Total=%Lu | SuccessRate=%Lf%% | FailureRate=%Lf%%\n",
			tid, res[tid]->success, res[tid]->failure, res[tid]->success+res[tid]->failure,
			PERCENT(res[tid]->success, loops), PERCENT(res[tid]->failure, loops));
        
        /* Give summary of all statues seen */
        unsigned long long  abort_explicit = 0,
                            abort_retry    = 0,
                            abort_conflict = 0,
                            abort_capacity = 0,
                            abort_debug    = 0,
                            abort_nested   = 0;

        #define CHECK_AND_INCR(status,flag,variable) \
                    if ((status)&(flag)) (variable)++

        for (int run = 0; run < loops; run++) {
            auto status = res[tid]->statuses[run];

            CHECK_AND_INCR(status, _XABORT_EXPLICIT, abort_explicit);
            CHECK_AND_INCR(status, _XABORT_RETRY, abort_retry);
            CHECK_AND_INCR(status, _XABORT_CONFLICT, abort_conflict);
            CHECK_AND_INCR(status, _XABORT_CAPACITY, abort_capacity);
            CHECK_AND_INCR(status, _XABORT_DEBUG, abort_debug);
            CHECK_AND_INCR(status, _XABORT_NESTED, abort_nested);
        }

        printf("      - Aborts: explicit=%Ld | retry=%Ld | conflict=%Ld | capacity=%Ld | debug=%Ld | nested=%Ld\n",
                abort_explicit,
                abort_retry,
                abort_conflict,
                abort_capacity,
                abort_debug,
                abort_nested);
        printf("\n");
	}

    /* Display overall stats */
	printf("All threads managed to count to %Ld\n", counter);

	printf("Total Attempts %Lu | Total Success %Lu | Total Failed %Lu\n", (unsigned long long)(total), success, failure);
	printf("Success Rate %Lu / %Lu = %Lf%%\n", success, (unsigned long long)(total), PERCENT(success, total));
	printf("Failure Rate %Lu / %Lu = %Lf%%\n", failure, (unsigned long long)(total), PERCENT(failure, total) );
	cout << "It took " << duration_cast<duration<double>>(stop - start).count() << " s" << endl;
	return 0;
}
